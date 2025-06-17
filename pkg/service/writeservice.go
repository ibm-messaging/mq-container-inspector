package service

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// WriteServiceYamlsToFile writes each service's details to a separate YAML file.
// Parameters:
//   - serviceList:        the list of services whose details will be written.
//   - fileNameFormat:     the format string used to name each file; must contain one "%s", which will be replaced by the service name.
//   - outputDir:          the directory in which the YAML files will be created.
func WriteServiceYamlsToFile(serviceList []corev1.Service, filNameFormat, outputDir string) error {

	for _, service := range serviceList {

		fileName := utils.FormatFilePath(outputDir, filNameFormat, service.Name)

		data, err := yaml.Marshal(service)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for service %s: %v", service.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing service %s data in the yaml file: %v", service.Name, err)
		}

	}

	return nil

}
