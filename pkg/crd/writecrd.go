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

package crd

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"sigs.k8s.io/yaml"
)

// WriteCRDYamlsToFile writes the CRD details to a it's YAML file.
// Parameters:
//   - crdDetails:             the CRD details.
//   - fileNameFormat:         the format string used to name each file
//   - outputDir:              the directory in which the YAML files will be created.
func WriteCRDYamlsToFile(crdDetails map[string]interface{}, fileNameFormat, outputDir string) error {

	fileName := utils.FormatFilePath(outputDir, fileNameFormat)

	data, err := yaml.Marshal(crdDetails)
	if err != nil {
		return fmt.Errorf("error while marshalling yaml for queue manager CRD: %v", err)
	}

	if err := os.WriteFile(fileName, data, 0660); err != nil {
		return fmt.Errorf("error while writing queue manager CRD data to yaml file %s: %v", fileName, err)
	}

	return nil
}
