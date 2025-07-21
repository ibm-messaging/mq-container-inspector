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
package csv

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// WriteCSVYamlsToFile writes each operator csv details to a separate YAML file.
// Parameters:
//   - csvDetailsList: the list of csv's whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the csv name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteCSVYamlsToFile(csvDetailsList []unstructured.Unstructured, fileNameFormat, outputDir string) error {

	for _, csvObj := range csvDetailsList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, csvObj.GetName())

		data, err := yaml.Marshal(csvObj)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for csv %s: %v", csvObj.GetName(), err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing csv %s data in the yaml file: %v", csvObj.GetName(), err)
		}

	}

	return nil

}
