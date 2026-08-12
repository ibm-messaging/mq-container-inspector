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

package validations

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/cr"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pvcinspector"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
)

// vrmfReleasePattern matches a bare "V.R.M.F-release" version string
var vrmfReleasePattern = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+-[A-Za-z0-9]+$`)

func ValidateQueueManagerImage(dynamicClient dynamic.Interface, flags utils.PVCInspectorFlags, logger *slog.Logger) (pvcinspector.QMImageInfo, error) {
	// check if the `--qm-image` flag is populated
	if flags.QueueManagerImage == "" {
		return resolveImageFromQueueManagerCR(dynamicClient, flags, logger)
	}

	var info pvcinspector.QMImageInfo

	// A full image URL always has a registry/repo path component, i.e. contains "/"
	if strings.Contains(flags.QueueManagerImage, "/") {
		info.Image = flags.QueueManagerImage
		// Best-effort: if the tag happens to be VRMF-shaped, capture it so
		// compareQueueManagerVersion can check it against the CR.
		if tag, ok := extractVRMFTag(flags.QueueManagerImage); ok {
			info.Version = tag
		}
	} else if vrmfReleasePattern.MatchString(flags.QueueManagerImage) {
		info.Version = flags.QueueManagerImage
		info.Image = fmt.Sprintf("%s/%s:%s", utils.DefaultQMImageRegistry, utils.DefaultQMImageType, flags.QueueManagerImage)
	} else {
		return pvcinspector.QMImageInfo{}, fmt.Errorf("invalid --qm-image value %q: must be a full image URL or a V.R.M.F-release version", flags.QueueManagerImage)
	}

	if flags.QueueManagerName != "" {
		qm, err := cr.GetQueueManagerCrDetailsByName(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
		switch {
		case err == nil:
			if updateErr := updateFromQueueManagerCR(qm, &info, flags, logger); updateErr != nil {
				return pvcinspector.QMImageInfo{}, updateErr
			}
		case errors.IsNotFound(err):
			logger.Info(fmt.Sprintf("QueueManager CR %s not found in namespace %s; proceeding without CR cross-check", flags.QueueManagerName, flags.QueueManagerNamespace))
		default:
			return pvcinspector.QMImageInfo{}, fmt.Errorf("error fetching QueueManager CR %s in namespace %s: %w", flags.QueueManagerName, flags.QueueManagerNamespace, err)
		}
	}

	return info, nil

}

func resolveImageFromQueueManagerCR(dynamicClient dynamic.Interface, flags utils.PVCInspectorFlags, logger *slog.Logger) (pvcinspector.QMImageInfo, error) {
	if flags.QueueManagerName == "" {
		return pvcinspector.QMImageInfo{}, fmt.Errorf("--qm-image was not provided and --qm-name is empty; cannot determine which image to use")
	}

	qm, err := cr.GetQueueManagerCrDetailsByName(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		return pvcinspector.QMImageInfo{}, fmt.Errorf("--qm-image was not provided; unable to fetch QueueManager CR %s in namespace %s to derive a version: %w", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	qmSpec, ok := qm["spec"].(map[string]interface{})
	if !ok {
		return pvcinspector.QMImageInfo{}, fmt.Errorf("unable to parse spec field from QueueManager CR %s in namespace %s: unexpected type or missing spec", flags.QueueManagerName, flags.QueueManagerNamespace)
	}

	version, ok := qmSpec["version"].(string)
	if !ok {
		return pvcinspector.QMImageInfo{}, fmt.Errorf("unable to parse spec.version field from QueueManager CR %s in namespace %s: unexpected type or missing field", flags.QueueManagerName, flags.QueueManagerNamespace)
	}

	logger.Info(fmt.Sprintf("--qm-image not provided; using version %q from QueueManager CR %s in namespace %s", version, flags.QueueManagerName, flags.QueueManagerNamespace))

	info := pvcinspector.QMImageInfo{
		Image:     fmt.Sprintf("%s/%s:%s", utils.DefaultQMImageRegistry, utils.DefaultQMImageType, version),
		Version:   version,
		QMCRFound: true,
	}
	info.StorageIsEphemeral = isStorageEphemeral(qmSpec)

	return info, nil
}

func extractVRMFTag(image string) (string, bool) {
	if strings.Contains(image, "@") {
		return "", false // digest-pinned, no tag to compare
	}

	lastSlash := strings.LastIndex(image, "/")
	lastColon := strings.LastIndex(image, ":")

	if lastColon == -1 || lastColon < lastSlash {
		return "", false // no tag present
	}

	tag := image[lastColon+1:]
	if !vrmfReleasePattern.MatchString(tag) {
		return "", false
	}

	return tag, true
}

func updateFromQueueManagerCR(qm map[string]interface{}, info *pvcinspector.QMImageInfo, flags utils.PVCInspectorFlags, logger *slog.Logger) error {
	qmSpec, ok := qm["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unable to parse spec field from QueueManager CR %s in namespace %s: unexpected type or missing spec", flags.QueueManagerName, flags.QueueManagerNamespace)
	}

	info.QMCRFound = true
	info.StorageIsEphemeral = isStorageEphemeral(qmSpec)

	qmSpecVersion, ok := qmSpec["version"].(string)
	if !ok {
		return fmt.Errorf("unable to parse version field from QueueManager CR %s in namespace %s: unexpected type or missing field", flags.QueueManagerName, flags.QueueManagerNamespace)
	}

	switch {
	case info.Version == "":
		logger.Info(fmt.Sprintf("No version could be determined from --qm-image; QueueManager CR %s in namespace %s reports version %q", flags.QueueManagerName, flags.QueueManagerNamespace, qmSpecVersion))
	case !strings.EqualFold(qmSpecVersion, info.Version):
		wasFullUrl := strings.Contains(flags.QueueManagerImage, "/")
		logger.Info(fmt.Sprintf("Version %q resolved from --qm-image does not match QueueManager CR %s version %q in namespace %s", info.Version, flags.QueueManagerName, qmSpecVersion, flags.QueueManagerNamespace))
		if !wasFullUrl {
			info.Version = qmSpecVersion
			info.Image = fmt.Sprintf("%s/%s:%s", utils.DefaultQMImageRegistry, utils.DefaultQMImageType, qmSpecVersion)
			logger.Info(fmt.Sprintf("Using QueueManager CR version %q instead", qmSpecVersion))
		}
	}

	return nil
}

func isStorageEphemeral(qmSpec map[string]interface{}) bool {
	qmSpecQM, ok := qmSpec["queueManager"].(map[string]interface{})
	if !ok {
		return false
	}

	qmSpecQMStorage, ok := qmSpecQM["storage"].(map[string]interface{})
	if !ok {
		return false
	}

	qmSpecQMStorageQM, ok := qmSpecQMStorage["queueManager"].(map[string]interface{})
	if !ok {
		return false
	}

	storageType, ok := qmSpecQMStorageQM["type"].(string)

	return ok && storageType == "ephemeral"
}
