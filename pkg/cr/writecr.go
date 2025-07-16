package cr

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
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
