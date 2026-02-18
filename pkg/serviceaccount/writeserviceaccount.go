/*
© Copyright IBM Corporation 2026

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

package serviceaccount

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
)

// WriteServiceAccountYamlsToFile writes each service-account details to a separate YAML file.
// Parameters:
//   - serviceAccountList:        the list of service-account's whose details will be written.
//   - fileNameFormat:     the format string used to name each file; must contain one "%s", which will be replaced by the service-account name.
//   - outputDir:          the directory in which the YAML files will be created.
func WriteServiceAccountYamlsToFile(serviceAccountList []corev1.ServiceAccount, filNameFormat, outputDir string) error {

	for _, serviceAccount := range serviceAccountList {

		fileName := utils.FormatFilePath(outputDir, filNameFormat, serviceAccount.ObjectMeta.Name)

		data, err := yaml.Marshal(serviceAccount)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for service-account %s: %v", serviceAccount.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return fmt.Errorf("error while writing service-account %s data in the yaml file: %v", serviceAccount.ObjectMeta.Name, err)
		}

	}

	return nil

}
