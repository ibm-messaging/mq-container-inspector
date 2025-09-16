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

package controllerrevisions

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

// WriteControllerRevisionYamlsToFile writes the Controller revisions as a list to a single file.
// Parameters:
//   - controllerRevisionList:  the list of controller revision whose details will be written.
//   - fileNameFormat:           the format string used to name each file; must contain one "%s", which will be replaced by the controller name.
//   - outputDir:                the directory in which the YAML files will be created.
func WriteControllerRevisionYamlsToFile(controllerRevisionList []appsv1.ControllerRevision, fileNameFormat, outputDir string) error {

	if len(controllerRevisionList) > 0 {

		statefulSetName := controllerRevisionList[0].OwnerReferences[0].Name

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, statefulSetName)

		// creating a Kind:List, object to store all the revisions in a single file
		revisionListMap := map[string]interface{}{
			"apiVersion": utils.ApiVersionV1,
			"kind":       utils.KindList,
			"items":      controllerRevisionList,
			"metadata": map[string]interface{}{
				"resourceVersion": "",
			},
		}

		data, err := yaml.Marshal(revisionListMap)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for StatefulSet revision %s: %v", statefulSetName, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing StatefulSet revision %s data in the yaml file: %v", statefulSetName, err)
		}
	}

	return nil
}
