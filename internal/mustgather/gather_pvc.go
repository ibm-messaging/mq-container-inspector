/*
© Copyright IBM Corporation 2025

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

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pvc"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

func gatherPVCToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create pvc directory to store pvc files
	pvcDirectory := filepath.Join(flags.OutputDir, "pvcs")
	directoryExist := utils.CheckIfDirectoryExist(pvcDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(pvcDirectory, 0775); err != nil {
			return err
		}
	}

	var pvcList []corev1.PersistentVolumeClaim
	var err error

	if flags.QueueManagerName != "" {
		pvcLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

		// get pvc's by selector
		pvcList, err = pvc.GetPVCDetailsBySelector(coreClient, pvcLabelSelector, flags.QueueManagerNamespace)
		if err != nil {
			logger.Info(fmt.Sprintf("fetching pvc's with selector %s: %v", pvcLabelSelector, err))
		}
	} else if flags.PodName != "" {

		// get the pvc's by pod name
		pvcList, err = pvc.GetPVCDetailsByPodName(coreClient, flags.PodName, flags.QueueManagerNamespace)
		if err != nil {
			logger.Info(fmt.Sprintf("fetching pvc's for %s pod: %v", flags.PodName, err))
		}

	}

	// write the pvc's to their respective yaml files
	pvcYamlFileNameFormat := "%s-pvc.yaml"
	if err := pvc.WritePVCYamlsToFile(pvcList, pvcYamlFileNameFormat, pvcDirectory); err != nil {
		logger.Error(err.Error())
	}

	if fileCount, err := utils.GetFileCountInDirectory(pvcDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("PVC Details: %s: Total Files: %d", pvcDirectory, fileCount))
	}

	return nil

}
