package ingress

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
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
