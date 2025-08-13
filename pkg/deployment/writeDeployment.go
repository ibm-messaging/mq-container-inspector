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
package deployment

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

// WriteDeploymentsToFile writes each deployment details to a separate YAML file.
// Parameters:
//   - deploymentList: the list of deployments whose details will be written.
//   - fileNameFormat: the format string used to name each file, must contain one "%s", which will be replaced by the deployment name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteDeploymentsToFile(deploymentList []appsv1.Deployment, fileNameFormat, outputDir string) error {

	for _, deployment := range deploymentList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, deployment.Name)

		data, err := yaml.Marshal(deployment)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for deployment %s: %v", deployment.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing %s deployment data in the yaml file: %v", deployment.Name, err)
		}

	}

	return nil

}

// WriteDeploymentEventsToFile writes each deployment events to a separate file.
// Parameters:
//   - deploymentsEventMap: the map of deployment names to their events, which will be written
//   - fileNameFormat: the format string used to name each file, must contain one "%s", which will be replaced by the deployment name.
//   - outputDir:      the directory in which the event files will be created.
func WriteDeploymentEventsToFile(deploymentsEventMap map[string][]corev1.Event, fileNameFormat, outputDir string) error {

	for deploymentName, deploymentEvents := range deploymentsEventMap {

		if len(deploymentEvents) > 0 {

			fileName := utils.FormatFilePath(outputDir, fileNameFormat, deploymentName)

			file, err := os.Create(fileName)
			if err != nil {
				return fmt.Errorf("error creating %s deployment event file %s: %v", deploymentName, fileName, err)
			}

			// tabwriter will handle dynamic spacing
			writer := tabwriter.NewWriter(file, 0, 8, 2, ' ', 0)

			// write the header
			fmt.Fprintf(writer, "LAST SEEN\tTYPE\tREASON\tKIND\tNAME\tMESSAGE\n")

			for _, event := range deploymentEvents {

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
