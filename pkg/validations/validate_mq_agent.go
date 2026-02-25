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

package validations

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/tabwriter"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/deployment"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
)

func ValidateMQAgentReleaseName(coreClient kubernetes.Interface, deploymentList []appsv1.Deployment, agentReleaseName, namespace string, logger *slog.Logger) error {

	if len(deploymentList) != 0 {
		return nil
	}

	logger.Info(fmt.Sprintf("No deployments found for %s mq-agent release name in %s namespace", agentReleaseName, namespace))
	fmt.Printf("No mq-agent deployments were found for release name '%s' in namespace '%s'.\n", agentReleaseName, namespace)
	// fetch all the deployments in the namespace with MQ Agent managed by label
	deploymentList, err := deployment.GetDeploymentsBySelector(coreClient, utils.MQAgentManagedByLabel, namespace)
	if err != nil {
		logger.Error(fmt.Sprintf("Error fetching deployments in the %s namespace: %v", namespace, err))
		return fmt.Errorf("'%s' is not a valid mq-agent release name. Please enter a valid release name.", agentReleaseName)
	}

	deploymentReleaseNameMap := make(map[string]string)

	for _, deployment := range deploymentList {

		annotations := deployment.Annotations

		if value, ok := annotations[utils.HelmReleaseNameAnnotation]; ok && value != "" {
			if strings.HasPrefix(deployment.ObjectMeta.Name, utils.MQAgentResourceNamePrefix) {
				deploymentReleaseNameMap[deployment.ObjectMeta.Name] = value
			}
		}
	}

	if len(deploymentReleaseNameMap) == 0 {
		fmt.Printf("No mq-agent deployments were found in namespace '%s'\n", namespace)
		return fmt.Errorf("invalid mq-agent release name '%s': no mq-agent deployments exist in namespace '%s'", agentReleaseName, namespace)
	}

	// print all the release names in the namespace
	fmt.Printf("Available mq-agent release names in %s namespace:\n", namespace)
	printMQAgentReleaseDetails(deploymentReleaseNameMap)

	return fmt.Errorf("no deployments found for %s mq-agent release name in %s namespace", agentReleaseName, namespace)

}

func printMQAgentReleaseDetails(deploymentReleaseNameMap map[string]string) {

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print Header
	fmt.Fprintln(writer, "RELEASE_NAME\tDEPLOYMENT_NAME")

	for deploymentName, releaseName := range deploymentReleaseNameMap {
		fmt.Fprintf(writer, "%s\t%s\n", releaseName, deploymentName)
	}

	writer.Flush()

}
