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
