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
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetDeploymentBySelector retrieves all deployments in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the deployments.
//   - namespace: the namespace in which to search for the deployments.
func GetDeploymentsBySelector(client kubernetes.Interface, selector, namespace string) ([]appsv1.Deployment, error) {

	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the Deployemt list API returns the deployments with the apiVersion and kind field as empty, so setting them explicitly
	for index := range deploymentList.Items {
		deploymentList.Items[index].TypeMeta.APIVersion = utils.ApiVersionAppsV1
		deploymentList.Items[index].TypeMeta.Kind = utils.KindDeployment
	}

	return deploymentList.Items, err

}

// GetDeploymentEventsBySelector retrieves all depoyment events in a given namespace that match the provided selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the deployments.
//   - namespace: the namespace in which to search for the deployments.
func GetDeploymentEventsBySelector(client kubernetes.Interface, selector, namespace string) (map[string][]corev1.Event, error) {

	deployments, err := GetDeploymentsBySelector(client, selector, namespace)
	if err != nil {
		return nil, err
	}

	deploymentEvents := make(map[string][]corev1.Event)

	for _, deployment := range deployments {

		fieldSelector := fmt.Sprintf("involvedObject.name=%s", deployment.ObjectMeta.Name)

		eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return nil, err
		}

		deploymentEvents[deployment.Name] = eventList.Items

	}

	return deploymentEvents, nil

}

// GetDeploymentsByPodName retrieves all deployments in a given namespace that match the provided pod name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - podName: the pod name used to filter the deployments.
//   - namespace: the namespace in which to search for the deployments.
func GetDeploymentsByPodName(client kubernetes.Interface, podName, namesapce string) (*appsv1.Deployment, error) {

	pod, err := pods.GetPodByName(client, podName, namesapce)
	if err != nil {
		return nil, err
	}

	// get the pod's controller owner name
	ownerName := ""
	for _, owner := range pod.ObjectMeta.OwnerReferences {
		if *owner.Controller {
			ownerName = owner.Name
			break
		}
	}
	if ownerName == "" {
		return nil, fmt.Errorf("error no replica set found as the contoller owner for the %s pod", podName)
	}

	// Get the deployment name from the replicaSet
	replicaSet, err := client.AppsV1().ReplicaSets(namesapce).Get(context.TODO(), ownerName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	deploymentName := ""
	for _, owner := range replicaSet.ObjectMeta.OwnerReferences {
		if *owner.Controller {
			deploymentName = owner.Name
			break
		}
	}
	if deploymentName == "" {
		return nil, fmt.Errorf("error no deployemnt found as the contoller owner for the %s replica set", replicaSet.Name)
	}

	deployment, err := client.AppsV1().Deployments(namesapce).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// the Deployemt get API returns the deployments with the apiVersion and kind field as empty, so setting them explicitly
	deployment.TypeMeta.APIVersion = utils.ApiVersionAppsV1
	deployment.TypeMeta.Kind = utils.KindDeployment

	return deployment, nil

}

// GetDeploymentEventsByName retrieves all depoyment events in a given namespace that match the provided deployment name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - deploymentName: the deployment name to fetch the events for.
//   - namespace: the namespace in which to search for the deployments.
func GetDeploymentEventsByName(client kubernetes.Interface, deploymentName, namespace string) (map[string][]corev1.Event, error) {

	deploymentEvents := make(map[string][]corev1.Event)

	fieldSelector := fmt.Sprintf("involvedObject.name=%s", deploymentName)

	eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	deploymentEvents[deploymentName] = eventList.Items

	return deploymentEvents, nil
}
