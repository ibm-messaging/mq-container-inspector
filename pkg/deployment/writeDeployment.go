package deployment

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

// WriteDeploymentsToFile writes each mq-operator deployment details to a separate YAML file.
// Parameters:
//   - deploymentList: the list of deployments whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the deployment name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteDeploymentsToFile(deploymentList []appsv1.Deployment, fileNameFormat, outputDir string) error {

	for _, deployment := range deploymentList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, deployment.Name)

		data, err := yaml.Marshal(deployment)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for mq-operator deployment %s: %v", deployment.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing mq-operator pod %s data in the yaml file: %v", deployment.Name, err)
		}

	}

	return nil

}
