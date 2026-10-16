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

package pdb

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	policyv1 "k8s.io/api/policy/v1"
	"sigs.k8s.io/yaml"
)

// WritePdbYamlsToFile writes each PodDisruptionBudget's details to a separate YAML file.
// Parameters:
//   - pdbList:        the list of PodDisruptionBudgets whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the PodDisruptionBudget name.
//   - outputDir:      the directory in which the YAML files will be created.
func WritePdbYamlsToFile(pdbList []policyv1.PodDisruptionBudget, fileNameFormat, outputDir string) error {

	for _, pdb := range pdbList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, pdb.ObjectMeta.Name)

		data, err := yaml.Marshal(pdb)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for pdb %s: %v", pdb.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return fmt.Errorf("error while writing pdb %s data in the yaml file: %v", pdb.ObjectMeta.Name, err)
		}

	}

	return nil
}
