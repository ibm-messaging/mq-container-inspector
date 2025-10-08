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

package ingress

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
)

// WriteIngressYamlsToFile writes each ingress's details to a separate YAML file.
// Parameters:
//   - ingressList:        the list of ingress whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the ingress name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteIngressYamlsToFile(ingressList []networkingv1.Ingress, fileNameFormat, outputDir string) error {

	for _, ingress := range ingressList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, ingress.ObjectMeta.Name)

		data, err := yaml.Marshal(ingress)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for ingress %s: %v", ingress.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing ingress %s data in the yaml file: %v", ingress.ObjectMeta.Name, err)
		}

	}

	return nil
}
