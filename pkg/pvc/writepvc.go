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
package pvc

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// WritePVCYamlsToFile writes each pvc's details to a separate YAML file.
// Parameters:
//   - pvcList:        the list of pvc's whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the pvc name.
//   - outputDir:      the directory in which the YAML files will be created.
func WritePVCYamlsToFile(pvcList []corev1.PersistentVolumeClaim, fileNameFormat, outputDir string) error {

	for _, pvc := range pvcList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, pvc.Name)

		data, err := yaml.Marshal(pvc)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for pvc %s: %v", pvc.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing pvc %s data in the yaml file: %v", pvc.Name, err)
		}

	}

	return nil
}
