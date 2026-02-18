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

package replicaset

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"sigs.k8s.io/yaml"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WriteReplicaSetToFile writes each replicaset details to a separate YAML file.
// Parameters:
//   - replicasetList: the list of replicasets whose details will be written.
//   - fileNameFormat: the format string used to name each file, must contain one "%s", which will be replaced by the replicaset name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteReplicaSetToFile(replicasetList []appsv1.ReplicaSet, fileNameFormat, outputDir string) error {

	for _, replicaset := range replicasetList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, replicaset.ObjectMeta.Name)

		data, err := yaml.Marshal(replicaset)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for replicaset %s: %v", replicaset.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return fmt.Errorf("error while writing %s replicaset data in the yaml file: %v", replicaset.ObjectMeta.Name, err)
		}

	}

	return nil

}

// WriteReplicaSetEventsToFile writes each replicaset events to a separate file.
// Parameters:
//   - replicasetEventsMap: the map of replicaset names to their events, which will be written
//   - fileNameFormat: the format string used to name each file, must contain one "%s", which will be replaced by the replicaset name.
//   - outputDir:      the directory in which the event files will be created.
func WriteReplicaSetEventsToFile(replicasetEventsMap map[string][]corev1.Event, fileNameFormat, outputDir string) error {

	for replicasetName, replicasetEvent := range replicasetEventsMap {

		if len(replicasetEvent) > 0 {

			filePath := utils.FormatFilePath(outputDir, fileNameFormat, replicasetName)

			file, err := utils.SafeOpenFile(outputDir, filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
			if err != nil {
				return fmt.Errorf("error creating file(%s): %v", filePath, err)
			}

			// tabwriter will handle dynamic spacing
			writer := tabwriter.NewWriter(file, 0, 8, 2, ' ', 0)

			// write the header
			fmt.Fprintf(writer, "LAST SEEN\tTYPE\tREASON\tKIND\tNAME\tMESSAGE\n")

			for _, event := range replicasetEvent {

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
