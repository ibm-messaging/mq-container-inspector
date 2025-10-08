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
package cr

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// WriteQueueManagerCrdYamlToFiles writes the QueueManager details to a it's YAML file.
// Parameters:
//   - queueManagerDetailsMap: the QueueManager details.
//   - fileNameFormat:         the format string used to name each file; must contain one "%s", which will be replaced by the QueueManager name.
//   - outputDir:              the directory in which the YAML files will be created.
func WriteQueueManagerCrdYamlToFiles(queueManagerDetailsMap map[string]interface{}, fileNameFormat, outputDir, queueManagerName string) error {

	fileName := utils.FormatFilePath(outputDir, fileNameFormat, queueManagerName)

	data, err := yaml.Marshal(queueManagerDetailsMap)
	if err != nil {
		return fmt.Errorf("error while marshalling yaml for queue manager %s: %v", queueManagerName, err)
	}

	if err := os.WriteFile(fileName, data, 0660); err != nil {
		return fmt.Errorf("error while writing queue manager %s data to yaml file %s: %v", queueManagerName, fileName, err)
	}

	return nil

}

// WriteIntegrationKeycloakClientCrdYamlToFiles writes the IntegrationKeycloakClient details to a it's YAML file.
// Parameters:
//   - integrationKeycloakClientDetailsList: the IntegrationKeycloakClient details.
//   - fileNameFormat:         the format string used to name each file; must contain one "%s", which will be replaced by the IntegrationKeycloakClient name.
//   - outputDir:              the directory in which the YAML files will be created.
func WriteIntegrationKeycloakClientCrdYamlToFiles(integrationKeycloakClientDetailsList []*unstructured.Unstructured, fileNameFormat, outputDir string) error {

	for _, integrationKeycloakClient := range integrationKeycloakClientDetailsList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, integrationKeycloakClient.GetName())

		data, err := yaml.Marshal(integrationKeycloakClient.Object)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for inetgration-keycloak-client %s: %v", integrationKeycloakClient.GetName(), err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing queue manager %s data to yaml file %s: %v", integrationKeycloakClient.GetName(), fileName, err)
		}

	}

	return nil

}
