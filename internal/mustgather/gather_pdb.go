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

package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pdb"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func gatherPDBToFiles(client kubernetes.Interface, flags utils.MustGatherFlags, qmPod *corev1.Pod, logger *slog.Logger) error {

	// create pdb directory to store pdb files
	pdbDirectory := filepath.Join(flags.OutputDir, "pdb")
	if !utils.CheckIfDirectoryExist(pdbDirectory) {
		if err := utils.CreateDirectory(pdbDirectory, 0o775); err != nil {
			return err
		}
	}

	var pdbList []policyv1.PodDisruptionBudget

	if qmPod != nil {
		pdbList = getPdbDetailsByPod(client, qmPod, flags.QueueManagerNamespace, logger)
	}

	if len(pdbList) == 0 && flags.QueueManagerName != "" {
		pdbList = getPdbDetailsByOwnerReference(client, flags.QueueManagerName, flags.QueueManagerNamespace, logger)
	}

	if len(pdbList) == 0 {
		logger.Info(fmt.Sprintf("PDB is enabled for %s but no PodDisruptionBudget resource was discovered in %s namespace",
			utils.QMIdentifier(qmPod, flags.QueueManagerName, flags.QueueManagerNamespace), flags.QueueManagerNamespace))
	}

	pdbFileNameFormat := "%s-pdb.yaml"
	if err := pdb.WritePdbYamlsToFile(pdbList, pdbFileNameFormat, pdbDirectory); err != nil {
		return err
	}

	if fileCount, err := utils.GetFileCountInDirectory(pdbDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("PDB details: %s: Total Files: %d", pdbDirectory, fileCount))
	}

	return nil

}

func getPdbDetailsByOwnerReference(client kubernetes.Interface, qmName, namespace string, logger *slog.Logger) []policyv1.PodDisruptionBudget {

	pdbLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", qmName)

	pdbList, err := pdb.GetPdbDetailsBySelector(client, pdbLabelSelector, namespace)

	if err != nil {
		logger.Error(fmt.Sprintf("error collecting pdb details for %s label selector in %s namespace: %v", pdbLabelSelector, namespace, err))
		return []policyv1.PodDisruptionBudget{}
	}

	var filteredPdbs []policyv1.PodDisruptionBudget

	// filter the list by owner reference
	for _, pdb := range pdbList {

		var ownerQMName string

		for _, owner := range pdb.ObjectMeta.OwnerReferences {
			if owner.Kind == utils.KindQueueManager && owner.Controller != nil && *owner.Controller {
				ownerQMName = owner.Name
				break
			}
		}

		if ownerQMName != "" && ownerQMName == qmName {
			filteredPdbs = append(filteredPdbs, pdb)
		}
	}

	return filteredPdbs

}

func getPdbDetailsByPod(client kubernetes.Interface, pod *corev1.Pod, namespace string, logger *slog.Logger) []policyv1.PodDisruptionBudget {

	pdbList, err := pdb.GetPdbDetails(client, namespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error collecting pdb details for %s pod in %s namespace: %v", pod.ObjectMeta.Name, namespace, err))
		return []policyv1.PodDisruptionBudget{}
	}

	var matchedPdbs []policyv1.PodDisruptionBudget

	for _, pdb := range pdbList {

		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			logger.Error(fmt.Sprintf("skipping PDB %s in %s namespace due to invalid selector: %v", pdb.ObjectMeta.Name, pdb.ObjectMeta.Namespace, err))
			continue
		}

		if selector.Matches(labels.Set(pod.Labels)) {
			matchedPdbs = append(matchedPdbs, pdb)
		}
	}

	return matchedPdbs

}
