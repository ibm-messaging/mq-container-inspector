package configmap

import (
	"fmt"
	"log/slog"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// WriteConfigMapYamlsToFile writes each configmap's details to a separate YAML file.
// Parameters:
//   - configMapList:        the list of configmap's whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the configmap name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteConfigMapYamlsToFile(configMapList []*corev1.ConfigMap, outputDir, fileNameFormat string, logger *slog.Logger) {

	for _, configmap := range configMapList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, configmap.ObjectMeta.Name)

		data, err := yaml.Marshal(configmap)
		if err != nil {
			logger.Info(fmt.Sprintf("error while marshalling yaml for configmap %s: %v", configmap.ObjectMeta.Name, err))
			continue
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			logger.Info(fmt.Sprintf("error while writing configmap %s data in the yaml file: %v", configmap.ObjectMeta.Name, err))
			continue
		}

	}

}
