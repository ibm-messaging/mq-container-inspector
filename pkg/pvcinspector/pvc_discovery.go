/*
© Copyright IBM Corporation 2026

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pvcinspector

import (
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strconv"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pvc"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type RegexMatchFunc func(qmName string) map[string]*regexp.Regexp

// operatorPVCPrefixToVolSpec maps the Operator's default PVC name prefixes to
// the corresponding volume name and mount path in the PVC-inspector pod.
// "-shared" entries are the MultiInstance-only shared persisted-data/recovery-logs
// PVCs, which carry no per-pod ordinal.
var operatorPVCPrefixToVolSpec = map[string]volumeSpec{
	"data":                  dataVolumeSpec,          // /mnt/mqm, per-replica
	"persisted-data":        persistedDataVolumeSpec, // /mnt/mqm-data, per-replica (SingleInstance/NativeHA)
	"recovery-logs":         recoveryLogsVolumeSpec,  // /mnt/mqm-log, per-replica (SingleInstance/NativeHA)
	"persisted-data-shared": persistedDataVolumeSpec, // /mnt/mqm-data, shared (MultiInstance)
	"recovery-logs-shared":  recoveryLogsVolumeSpec,  // /mnt/mqm-log, shared (MultiInstance)
}

// buildOperatorPVCNamePatterns builds anchored regular expressions for the
// Operator's default PVC naming convention.
func buildOperatorPVCNamePatterns(qmName string) map[string]*regexp.Regexp {
	q := regexp.QuoteMeta(qmName)
	return map[string]*regexp.Regexp{
		"data":                  regexp.MustCompile(fmt.Sprintf(`^data-%s-ibm-mq-(\d+)$`, q)),
		"persisted-data":        regexp.MustCompile(fmt.Sprintf(`^persisted-data-%s-ibm-mq-(\d+)$`, q)),
		"recovery-logs":         regexp.MustCompile(fmt.Sprintf(`^recovery-logs-%s-ibm-mq-(\d+)$`, q)),
		"persisted-data-shared": regexp.MustCompile(fmt.Sprintf(`^%s-ibm-mq-persisted-data$`, q)),
		"recovery-logs-shared":  regexp.MustCompile(fmt.Sprintf(`^%s-ibm-mq-recovery-logs$`, q)),
	}
}

// pvcPrefixToVolSpec maps the default IBM MQ PVC name prefixes to the
// corresponding volume name and mount path in the PVC-inspector pod.
//
// These are default naming conventions as per MQ Helm. If a deployment customizes its PVC
// names, the PVC may be selected through its instance label but will not match
// these patterns and therefore will be skipped.
var pvcPrefixToVolSpec = map[string]volumeSpec{
	"qm":   dataVolumeSpec,          // /mnt/mqm
	"data": persistedDataVolumeSpec, // /mnt/mqm-data
	"log":  recoveryLogsVolumeSpec,  // /mnt/mqm-log
}

// buildPVCNamePatterns builds anchored regular expressions for the known
// default PVC naming conventions.
//
// The expressions are anchored to the supplied QueueManager instance name to
// prevent accidental matching of PVCs belonging to another deployment whose
// name merely contains the same substring.
// Parameters:
//   - qmName: the QueueManager instance name used to build the PVC name patterns.
func buildPVCNamePatterns(qmName string) map[string]*regexp.Regexp {
	quotedName := regexp.QuoteMeta(qmName)

	return map[string]*regexp.Regexp{
		"qm": regexp.MustCompile(
			fmt.Sprintf(`^qm-%s(?:-ibm-mq)?(?:-(\d+))?$`, quotedName),
		),
		"data": regexp.MustCompile(
			fmt.Sprintf(`^data-%s(?:-ibm-mq)?(?:-(\d+))?$`, quotedName),
		),
		"log": regexp.MustCompile(
			fmt.Sprintf(`^log-%s(?:-ibm-mq)?(?:-(\d+))?$`, quotedName),
		),
	}
}

// matchedPVC represents a PVC that matched one of the supported default naming
// conventions.
// Fields:
//   - prefix: the matched PVC name prefix.
//   - ordinal: the pod ordinal extracted from the PVC name, or nil for shared storage.
//   - pvcName: the full PVC name.
type matchedPVC struct {
	prefix  string
	ordinal *int
	pvcName string
}

// getPVCMountDataByDiscovery retrieves PVC volume and volume mount
// information for a QueueManager purely through label/name-based discovery —
// no CR involved. It tries the Operator's naming convention first, then the
// Helm chart's, since both use the same app.kubernetes.io/instance label and
// can't otherwise be distinguished without a CR to consult. This covers
// Operator-managed QMs, Helm-managed QMs, and orphaned PVCs left behind by a
// deleted QueueManager CR.
// Parameters:
//   - coreClient: the Kubernetes client used to retrieve PersistentVolumeClaims.
//   - flags: the PVC inspector flags containing the QueueManager details.
//   - logger: the logger used to record informational and warning messages.
func getPVCMountDataByDiscovery(coreClient kubernetes.Interface, flags utils.PVCInspectorFlags, logger *slog.Logger) (map[string]PodPVCMountData, error) {

	if flags.QueueManagerName == "" {
		return nil, fmt.Errorf("when falling back to --qm-image, the --qm-name flag cannot be empty")
	}

	pvcRefsByPod, err := discoverPVCRefsByPod(coreClient, flags, operatorPVCPrefixToVolSpec, buildOperatorPVCNamePatterns, logger)
	if err != nil {
		return nil, err
	}

	if len(pvcRefsByPod) == 0 {
		logger.Info(fmt.Sprintf("No PVCs matched the Operator naming convention for QueueManager %q in namespace %s; trying the Helm chart naming convention", flags.QueueManagerName, flags.QueueManagerNamespace))

		pvcRefsByPod, err = discoverPVCRefsByPod(coreClient, flags, pvcPrefixToVolSpec, buildPVCNamePatterns, logger)
		if err != nil {
			return nil, err
		}
	}

	if len(pvcRefsByPod) == 0 {
		return nil, nil
	}

	return buildPodPVCMountDataMap(
		coreClient,
		pvcRefsByPod,
		flags.QueueManagerNamespace,
		logger,
	), nil
}

// discoverPVCRefsByPod discovers supported PVCs for a QueueManager deployment
// and groups them by PVC-inspector pod, using the supplied naming convention
// (prefix -> volumeSpec map and pattern builder). PVCs are first selected
// using the QueueManager instance label and are then matched against the
// convention's expected prefixes.
//
// Returns (nil, nil) — not an error — when no PVCs match this particular
// convention, so callers can try an alternate convention before deciding
// discovery has genuinely failed.
// Parameters:
//   - coreClient: the Kubernetes client used to retrieve PersistentVolumeClaims.
//   - flags: the PVC inspector flags containing the QueueManager details.
//   - prefixToVolSpec: maps this convention's PVC name prefixes to volume specs.
//   - buildPatterns: builds this convention's anchored name-matching patterns.
//   - logger: the logger used to record informational and warning messages.
func discoverPVCRefsByPod(coreClient kubernetes.Interface, flags utils.PVCInspectorFlags, prefixToVolSpec map[string]volumeSpec, buildPatterns RegexMatchFunc, logger *slog.Logger) (map[string][]pvcRef, error) {

	pvcList, err := listInstancePVCs(coreClient, flags)
	if err != nil {
		return nil, err
	}

	if len(pvcList) == 0 {
		logger.Info(fmt.Sprintf(
			"No PVCs found with label app.kubernetes.io/instance=%s in namespace %s; "+
				"skipping name-only discovery to avoid matching PVCs belonging to another deployment",
			flags.QueueManagerName,
			flags.QueueManagerNamespace,
		))

		return nil, nil
	}

	patterns := buildPatterns(flags.QueueManagerName)

	var matched []matchedPVC
	var unmatchedNames []string

	ordinalsSeen := make(map[int]struct{})

	for _, pvcItem := range pvcList {
		pvcMatched := false

		for prefix, pattern := range patterns {
			matches := pattern.FindStringSubmatch(pvcItem.Name)
			if matches == nil {
				continue
			}

			var ordinal *int

			if len(matches) > 1 && matches[1] != "" {
				value, err := strconv.Atoi(matches[1])
				if err != nil {
					logger.Warn(fmt.Sprintf(
						"Skipping PVC %q because its ordinal %q could not be parsed: %v",
						pvcItem.Name,
						matches[1],
						err,
					))
					continue
				}

				ordinal = &value
				ordinalsSeen[value] = struct{}{}
			}

			matched = append(matched, matchedPVC{
				prefix:  prefix,
				ordinal: ordinal,
				pvcName: pvcItem.Name,
			})

			pvcMatched = true
			break
		}

		if !pvcMatched {
			unmatchedNames = append(unmatchedNames, pvcItem.Name)
		}
	}

	if len(unmatchedNames) > 0 {
		logger.Info(fmt.Sprintf(
			"Skipping %d PVC(s) in namespace %s that matched QueueManager instance %q through labels but did not match a supported PVC naming convention: %v; the PVC names may have been customized and these PVCs will not be inspected",
			len(unmatchedNames),
			flags.QueueManagerNamespace,
			flags.QueueManagerName,
			unmatchedNames,
		))
	}

	if len(matched) == 0 || len(ordinalsSeen) == 0 {
		return nil, nil
	}

	logger.Info(fmt.Sprintf(
		"Discovered %d pod ordinal(s) for QueueManager instance %q in namespace %s: %v",
		len(ordinalsSeen),
		flags.QueueManagerName,
		flags.QueueManagerNamespace,
		ordinalKeys(ordinalsSeen),
	))

	podPVCRefs := make(
		map[string][]pvcRef,
		len(ordinalsSeen),
	)

	for ordinal := range ordinalsSeen {
		podName := fmt.Sprintf(
			"pvc-inspector-%s-ibm-mq-%d",
			flags.QueueManagerName,
			ordinal,
		)

		podPVCRefs[podName] = nil
	}

	for _, matchedPVC := range matched {
		volSpec, ok := prefixToVolSpec[matchedPVC.prefix]
		if !ok {
			logger.Warn(fmt.Sprintf(
				"Skipping PVC %q because prefix %q does not have a configured volume specification",
				matchedPVC.pvcName,
				matchedPVC.prefix,
			))
			continue
		}

		ref := pvcRef{
			volSpec: volSpec,
			pvcName: matchedPVC.pvcName,
		}

		if matchedPVC.ordinal != nil {
			podName := fmt.Sprintf(
				"pvc-inspector-%s-ibm-mq-%d",
				flags.QueueManagerName,
				*matchedPVC.ordinal,
			)

			podPVCRefs[podName] = append(
				podPVCRefs[podName],
				ref,
			)

			continue
		}

		// no ordinal in the name -> shared storage (e.g. MultiInstance), attach to every discovered pod
		logger.Info(fmt.Sprintf(
			"PVC %q has no ordinal suffix; treating it as shared storage and attaching it to all %d discovered PVC-inspector pod(s)",
			matchedPVC.pvcName,
			len(podPVCRefs),
		))

		for podName := range podPVCRefs {
			podPVCRefs[podName] = append(
				podPVCRefs[podName],
				ref,
			)
		}
	}

	return podPVCRefs, nil
}

// listInstancePVCs retrieves all PVCs in the configured namespace that match the
// QueueManager instance label.
// Parameters:
//   - coreClient: the Kubernetes client used to retrieve PersistentVolumeClaims.
//   - flags: the PVC inspector flags containing the QueueManager name and namespace.
func listInstancePVCs(coreClient kubernetes.Interface, flags utils.PVCInspectorFlags) ([]corev1.PersistentVolumeClaim, error) {

	selector := fmt.Sprintf(
		"app.kubernetes.io/instance=%s",
		flags.QueueManagerName,
	)

	pvcList, err := pvc.GetPVCDetailsBySelector(
		coreClient,
		selector,
		flags.QueueManagerNamespace,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error listing PVCs by instance label for %q in namespace %q: %w",
			flags.QueueManagerName,
			flags.QueueManagerNamespace,
			err,
		)
	}

	return pvcList, nil
}

// ordinalKeys returns the pod ordinals contained in the supplied map in sorted order.
// Parameters:
//   - ordinals: the set of discovered pod ordinals.
func ordinalKeys(ordinals map[int]struct{}) []int {
	keys := make([]int, 0, len(ordinals))

	for ordinal := range ordinals {
		keys = append(keys, ordinal)
	}

	sort.Ints(keys)

	return keys
}
