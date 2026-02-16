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

package mqagent

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"

	routeV1 "github.com/openshift/api/route/v1"
	routeClient "github.com/openshift/client-go/route/clientset/versioned"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/deployment"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/networkpolicy"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/replicaset"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/routes"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/service"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func CollectMQAgentDeploymentDetails(coreClient kubernetes.Interface, flags utils.MQAgentFlags, logger *slog.Logger) ([]string, error) {

	deploymentDirectory := filepath.Join(flags.OutputDir, "deployment")
	if !utils.CheckIfDirectoryExist(deploymentDirectory) {
		if err := utils.CreateDirectory(deploymentDirectory, 0o775); err != nil {
			return nil, err
		}
	}

	// collect deployment details
	deploymentList, err := deployment.GetDeploymentsBySelector(coreClient, utils.MQAgentLabels, flags.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error fetching deployment details in namespace %s: %v", flags.Namespace, err)
	}

	deploymentList = utils.FilterDeploymentListByAnnotation(deploymentList, utils.HelmReleaseNameAnnotation, flags.AgentReleaseName)
	if len(deploymentList) == 0 {
		logger.Error(fmt.Sprintf("No mq-agent deployment found for release-name %s in namespace %s", flags.AgentReleaseName, flags.Namespace))
		return []string{}, nil
	}

	// collect deployment events
	deploymentEvents, err := deployment.GetEventsForDeploymentList(coreClient, deploymentList)
	if err != nil {
		return nil, fmt.Errorf("error fetching deployment events in namespace %s: %v", flags.Namespace, err)
	}

	// write the deployments in their respective yaml files
	fileNameFormat := "%s-deployment.yaml"
	if err := deployment.WriteDeploymentsToFile(deploymentList, fileNameFormat, deploymentDirectory); err != nil {
		return nil, err
	}

	// write the deployment events in their respective files
	fileNameFormat = "%s-deployment-events.txt"
	if err := deployment.WriteDeploymentEventsToFile(deploymentEvents, fileNameFormat, deploymentDirectory); err != nil {
		return nil, err
	}

	// Check if the directory is empty
	if fileCount, err := utils.GetFileCountInDirectory(deploymentDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Deployment details: %s: Total Files: %d", deploymentDirectory, fileCount))
	}

	var deploymentNameList []string
	for _, deployment := range deploymentList {
		deploymentNameList = append(deploymentNameList, deployment.ObjectMeta.Name)
	}

	return deploymentNameList, nil

}

func CollectMQAgentReplicaSetDetails(coreClient kubernetes.Interface, flags utils.MQAgentFlags, deploymentNameList []string, logger *slog.Logger) ([]string, error) {

	if len(deploymentNameList) == 0 {
		logger.Error(fmt.Sprintf("Skipping ReplicaSet collection in %s namespace as the deployment list is empty", flags.Namespace))
		return []string{}, nil
	}

	replicaSetDirectory := filepath.Join(flags.OutputDir, "replicaset")
	if !utils.CheckIfDirectoryExist(replicaSetDirectory) {
		if err := utils.CreateDirectory(replicaSetDirectory, 0o775); err != nil {
			return nil, err
		}
	}

	// collect replicaset details
	replicaSetList, err := replicaset.GetReplicaSetBySelector(coreClient, utils.MQAgentLabels, flags.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error fetching replicaset details in namespace %s: %v", flags.Namespace, err)
	}

	replicaSetList = utils.FilterReplicasetListByAnnotation(replicaSetList, utils.HelmReleaseNameAnnotation, flags.AgentReleaseName)
	if len(replicaSetList) == 0 {
		logger.Error(fmt.Sprintf("No mq-agent replicaset found for release-name %s in namespace %s", flags.AgentReleaseName, flags.Namespace))
		return []string{}, nil
	}

	// remove any replicaSets which are not owned by a deployment in the deploymentList
	replicaSetList = filterReplicasetListByOwnerReferences(replicaSetList, deploymentNameList)

	// collect replicaSet events
	replicasetEvents, err := replicaset.GetEventsForReplicaSetList(coreClient, replicaSetList)
	if err != nil {
		return nil, fmt.Errorf("error fetching replicaset events in namespace %s: %v", flags.Namespace, err)
	}

	// write the replicasets in their respective yaml files
	fileNameFormat := "%s-replicaset.yaml"
	if err := replicaset.WriteReplicaSetToFile(replicaSetList, fileNameFormat, replicaSetDirectory); err != nil {
		return nil, err
	}

	// write the replicaset events in their respective files
	fileNameFormat = "%s-replicaset-events.txt"
	if err := replicaset.WriteReplicaSetEventsToFile(replicasetEvents, fileNameFormat, replicaSetDirectory); err != nil {
		return nil, err
	}

	// check if the directory is empty
	if fileCount, err := utils.GetFileCountInDirectory(replicaSetDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Replicaset details: %s: Total Files: %d", replicaSetDirectory, fileCount))
	}

	var replicasetNameList []string
	for _, replicaset := range replicaSetList {
		replicasetNameList = append(replicasetNameList, replicaset.ObjectMeta.Name)
	}

	return replicasetNameList, nil
}

func filterReplicasetListByOwnerReferences(replicaSetList []appsv1.ReplicaSet, deploymentNameList []string) []appsv1.ReplicaSet {

	var replicasets []appsv1.ReplicaSet

	for _, replicaset := range replicaSetList {

		var ownerName string
		var ownerKind string

		for _, owner := range replicaset.OwnerReferences {
			if *owner.Controller {
				ownerName = owner.Name
				ownerKind = owner.Kind
				break
			}
		}

		if ownerKind == utils.KindDeployment && slices.Contains(deploymentNameList, ownerName) {
			replicasets = append(replicasets, replicaset)
		}
	}

	return replicasets

}

func CollectMQAgentPodDetails(coreClient kubernetes.Interface, flags utils.MQAgentFlags, replicasetNameList []string, logger *slog.Logger) error {

	if len(replicasetNameList) == 0 {
		logger.Error(fmt.Sprintf("Skipping Pod collection in %s namespace as the replicaset list is empty", flags.Namespace))
		return nil
	}

	podDirectory := filepath.Join(flags.OutputDir, "pods")
	if !utils.CheckIfDirectoryExist(podDirectory) {
		if err := utils.CreateDirectory(podDirectory, 0o775); err != nil {
			return err
		}
	}

	// collect the pods
	podList, err := pods.GetPodsBySelector(coreClient, utils.MQAgentLabels, flags.Namespace)
	if err != nil {
		return fmt.Errorf("error fetching pods details in namespace %s: %v", flags.Namespace, err)
	}

	podList = validatePodsOwnerReferences(podList, replicasetNameList)

	// fetch all the containers in the pod
	podContainerNamesMap := utils.FetchPodContainers(podList)

	// fetch the pod logs
	podLogsMap := pods.GetPodLogsForAllContainers(coreClient, podContainerNamesMap, flags.Namespace, logger)

	// fetch the pod events
	podEvents, err := pods.GetPodEvents(coreClient, podList, flags.Namespace)
	if err != nil {
		return fmt.Errorf("error fetching pods events in namespace %s: %v", flags.Namespace, err)
	}

	// fetch the pod describe logs
	podDescribeLogs, err := pods.GetPodDescribeLogs(coreClient, podList, flags.Namespace)
	if err != nil {
		return fmt.Errorf("error fetching pods describe logs in namespace %s: %v", flags.Namespace, err)
	}

	// write the pod details to its text file
	fileNameFormat := "pod-details.txt"
	if err := pods.WritePodDetailsToFile(podList, fileNameFormat, podDirectory); err != nil {
		return err
	}

	// write the pod details to their respective yamls
	fileNameFormat = "%s.yaml"
	if err := pods.WritePodYamlsToFile(podList, fileNameFormat, podDirectory); err != nil {
		return err
	}

	// write the pod logs to their respective files
	fileNameFormat = "[%s]-%s-%s-pod-log.txt"
	if err := pods.WritePodAllContainerLogsToFile(podLogsMap, fileNameFormat, podDirectory, logger); err != nil {
		return err
	}

	// write the pod events to their respective files
	fileNameFormat = "%s-pod-events.txt"
	if err := pods.WritePodEventsToFile(podEvents, fileNameFormat, podDirectory); err != nil {
		return err
	}

	// write the pod describe logs to their respective files
	fileNameFormat = "%s-describe-log.txt"
	if err := pods.WritePodDescribeLogsToFile(podDescribeLogs, fileNameFormat, podDirectory); err != nil {
		return err
	}

	// check if the directory is empty
	if fileCount, err := utils.GetFileCountInDirectory(podDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Replicaset details: %s: Total Files: %d", podDirectory, fileCount))
	}

	return nil
}

func validatePodsOwnerReferences(podList []corev1.Pod, replicaSetNameList []string) []corev1.Pod {

	var pods []corev1.Pod

	for _, pod := range podList {

		var ownerName string
		var ownerKind string

		for _, owner := range pod.OwnerReferences {
			if *owner.Controller {
				ownerName = owner.Name
				ownerKind = owner.Kind
				break
			}
		}

		if ownerKind == utils.KindReplicaSet && slices.Contains(replicaSetNameList, ownerName) {
			pods = append(pods, pod)
		}

	}

	return pods

}

func CollectMQAgentServiceDetails(coreClient kubernetes.Interface, flags utils.MQAgentFlags, logger *slog.Logger) ([]string, error) {

	serviceDirectory := filepath.Join(flags.OutputDir, "service")
	if !utils.CheckIfDirectoryExist(serviceDirectory) {
		if err := utils.CreateDirectory(serviceDirectory, 0o755); err != nil {
			return nil, err
		}
	}

	// collect the services list
	serviceList, err := service.GetServiceDetailsBySelector(coreClient, utils.MQAgentManagedByLabel, flags.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error fetching services in namespace %s: %v", flags.Namespace, err)
	}

	// filter by release name annotation
	serviceList = utils.FilterServiceListByAnnotation(serviceList, utils.HelmReleaseNameAnnotation, flags.AgentReleaseName)

	if len(serviceList) == 0 {
		logger.Error(fmt.Sprintf("No mq-agent service found for release-name %s in namespace %s", flags.AgentReleaseName, flags.Namespace))
		return []string{}, nil
	}

	//write service details in their respective yaml files
	fileNameFormat := "%s-service.yaml"
	if err := service.WriteServiceYamlsToFile(serviceList, fileNameFormat, serviceDirectory); err != nil {
		return nil, err
	}

	// check if the directory is empty
	if fileCount, err := utils.GetFileCountInDirectory(serviceDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Service details: %s: Total Files: %d", serviceDirectory, fileCount))
	}

	var serviceNameList []string
	for _, service := range serviceList {
		serviceNameList = append(serviceNameList, service.ObjectMeta.Name)
	}
	return serviceNameList, nil
}

func CollectMQAgentRouteDetails(routeClient routeClient.Interface, serviceNameList []string, flags utils.MQAgentFlags, logger *slog.Logger) error {

	if len(serviceNameList) == 0 {
		logger.Error(fmt.Sprintf("Skipping Route collection in %s namespace as the service list is empty", flags.Namespace))
		return nil
	}

	routeDirectory := filepath.Join(flags.OutputDir, "route")
	if !utils.CheckIfDirectoryExist(routeDirectory) {
		if err := utils.CreateDirectory(routeDirectory, 0o755); err != nil {
			return err
		}
	}

	// collect the route list
	routeList, err := routes.GetRouteDetailsBySelector(routeClient, utils.MQAgentManagedByLabel, flags.Namespace)
	if err != nil {
		return fmt.Errorf("error fetching routes in namespace %s: %v", flags.Namespace, err)
	}

	// filter by release name annotation
	routeList = utils.FilterRouteListByAnnotation(routeList, utils.HelmReleaseNameAnnotation, flags.AgentReleaseName)

	// filter routes to only include those pointing to the identified services earlier
	routeList = filterRoutesByServices(routeList, serviceNameList)

	if len(routeList) == 0 {
		logger.Error(fmt.Sprintf("No routes found in namespace %s", flags.Namespace))
		return nil
	}

	// write route details in their respective yaml files
	fileNameFormat := "%s-route.yaml"
	if err := routes.WriteRouteDetailsBySelectorToFile(routeList, fileNameFormat, routeDirectory); err != nil {
		return err
	}

	// check if the directory is empty
	if fileCount, err := utils.GetFileCountInDirectory(routeDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Route details: %s: Total Files: %d", routeDirectory, fileCount))
	}

	return nil
}

func CollectMQAgentNetworkPolicies(coreClient kubernetes.Interface, flags utils.MQAgentFlags, logger *slog.Logger) error {

	networkPolicyDirectory := filepath.Join(flags.OutputDir, "networkpolicy")

	if !utils.CheckIfDirectoryExist(networkPolicyDirectory) {
		if err := utils.CreateDirectory(networkPolicyDirectory, 0o775); err != nil {
			return err
		}
	}

	// collect the network policy list
	networkPolicyList, err := networkpolicy.GetNetworkPolicyBySelector(coreClient, utils.MQAgentManagedByLabel, flags.Namespace)
	if err != nil {
		return fmt.Errorf("error fetching network-policies in %s namespace: %v", flags.Namespace, err)
	}

	// filter by release name annotation
	networkPolicyList = utils.FilterNetworkPolicyListByAnnotation(networkPolicyList, utils.HelmReleaseNameAnnotation, flags.AgentReleaseName)

	if len(networkPolicyList) == 0 {
		logger.Error(fmt.Sprintf("No network-policies found in namespace %s", flags.Namespace))
		return nil
	}

	// write network-policy details in their respective yaml files
	fileNameFormat := "%s-network-policy.yaml"
	if err := networkpolicy.WriteNetworkPolicyToFile(networkPolicyList, fileNameFormat, networkPolicyDirectory); err != nil {
		return err
	}

	// check if the directory is empty
	if fileCount, err := utils.GetFileCountInDirectory(networkPolicyDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("NetworkPolicy details: %s: Total Files: %d", networkPolicyDirectory, fileCount))
	}

	return nil

}

func filterRoutesByServices(routeList []routeV1.Route, serviceNameList []string) []routeV1.Route {

	var filteredRoutes []routeV1.Route

	for _, route := range routeList {
		if route.Spec.To.Name != "" && slices.Contains(serviceNameList, route.Spec.To.Name) {
			filteredRoutes = append(filteredRoutes, route)
		}
	}

	return filteredRoutes

}
