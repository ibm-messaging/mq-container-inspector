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
package statefulset

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// WriteStatefulSetYamlsToFile writes each StatefulSet details to a separate YAML file.
// Parameters:
//   - statefulSetList:  the list of StatefulSet whose details will be written.
//   - fileNameFormat:   the format string used to name each file; must contain one "%s", which will be replaced by the StatefulSet name.
//   - outputDir:        the directory in which the YAML files will be created.
func WriteStatefulSetYamlsToFile(statefulSetList []appsv1.StatefulSet, fileNameFormat, outputDir string) error {

	for _, statefulSet := range statefulSetList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, statefulSet.Name)

		data, err := yaml.Marshal(statefulSet)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for StatefulSet %s: %v", statefulSet.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing StatefulSet %s data in the yaml file: %v", statefulSet.Name, err)
		}

	}

	return nil

}

// WriteStatefulSetRevisionYamlsToFile writes the StatefulSet revisions as a list to a single file.
// Parameters:
//   - statefulSetRevisionList:  the list of StatefulSet revision whose details will be written.
//   - fileNameFormat:           the format string used to name each file; must contain one "%s", which will be replaced by the StatefulSet name.
//   - outputDir:                the directory in which the YAML files will be created.
func WriteStatefulSetRevisionYamlsToFile(statefulSetRevisionList []appsv1.ControllerRevision, fileNameFormat, outputDir string) error {

	if len(statefulSetRevisionList) > 0 {

		statefulSetName := statefulSetRevisionList[0].OwnerReferences[0].Name

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, statefulSetName)

		// creating a Kind:List, object to store all the revisions in a single file
		revisionListMap := map[string]interface{}{
			"apiVersion": utils.ApiVersionV1,
			"kind":       utils.KindList,
			"items":      statefulSetRevisionList,
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

// WriteStatefulSetEventsToFile writes each StatefulSet events to their respective files.
// Parameters:
//   - statefulSetEventsMap:  the map of StatefulSet names to their events, which will be written..
//   - fileNameFormat:        the format string used to name each file; must contain one "%s", which will be replaced by the StatefulSet name.
//   - outputDir:             the directory in which the YAML files will be created.
func WriteStatefulSetEventsToFile(statefulSetEventsMap map[string][]corev1.Event, fileNameFormat, outputDir string) error {

	for statefulSetName, statefulSetEvents := range statefulSetEventsMap {

		if len(statefulSetEvents) > 0 {

			fileName := utils.FormatFilePath(outputDir, fileNameFormat, statefulSetName)

			file, err := os.Create(fileName)
			if err != nil {
				return fmt.Errorf("error creating %s StatefulSet event file %s: %v", statefulSetName, fileName, err)
			}

			// tabwriter will handle dynamic spacing
			writer := tabwriter.NewWriter(file, 0, 8, 2, ' ', 0)

			// write the header
			fmt.Fprintf(writer, "LAST SEEN\tTYPE\tREASON\tKIND\tNAME\tMESSAGE\n")

			for _, event := range statefulSetEvents {

				var lastSeen string
				if !event.LastTimestamp.IsZero() {
					lastSeen = event.LastTimestamp.Format(time.RFC3339)
				}

				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
					lastSeen,
					event.Type,
					event.Reason,
					event.InvolvedObject.Kind,
					event.InvolvedObject.Name,
					event.Message,
				)
			}

			writer.Flush()
			file.Close()

		}
	}

	return nil

}
