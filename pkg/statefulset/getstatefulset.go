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
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetStatefulSetDetailsBySelector retrieves all StatefulSets in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the StatefulSets.
//   - namespace: the namespace in which to search for the StatefulSets.
func GetStatefulSetDetailsBySelector(client kubernetes.Interface, selector, namespace string) ([]appsv1.StatefulSet, error) {

	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the StatefulSet list API returns the statefulsets with the apiVersion and kind field as empty, so setting them explicitly
	for index := range statefulSetList.Items {
		statefulSetList.Items[index].TypeMeta.APIVersion = utils.ApiVersionAppsV1
		statefulSetList.Items[index].TypeMeta.Kind = utils.KindStatefulSet
	}

	return statefulSetList.Items, err

}

// GetStatefulSetEventsBySelector retrieves all StatefulSet events in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the StatefulSets.
//   - namespace: the namespace in which to search for the StatefulSets.
func GetStatefulSetEventsBySelector(client kubernetes.Interface, selector, namespace string) (map[string][]corev1.Event, error) {

	statefulSetList, err := GetStatefulSetDetailsBySelector(client, selector, namespace)
	if err != nil {
		return nil, err
	}

	statefulSetEventsMap := make(map[string][]corev1.Event)

	for _, statefulSet := range statefulSetList {

		fieldSelector := fmt.Sprintf("involvedObject.name=%s", statefulSet.Name)

		statefulSetEvents, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return statefulSetEventsMap, err
		}

		statefulSetEventsMap[statefulSet.Name] = statefulSetEvents.Items

	}

	return statefulSetEventsMap, nil

}

// GetStatefulSetDetailsByPodName retrieves all StatefulSet details in a given namespace that match the provided pod name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - podName: the pod name used to filter the StatefulSets.
//   - namespace: the namespace in which to search for the StatefulSets.
func GetStatefulSetDetailsByPodName(client kubernetes.Interface, podName, namespace string) (*appsv1.StatefulSet, error) {

	pod, err := pods.GetPodByName(client, podName, namespace)
	if err != nil {
		return nil, err
	}

	ownerName := ""
	for _, owner := range pod.ObjectMeta.OwnerReferences {
		if *owner.Controller {
			ownerName = owner.Name
			break
		}
	}
	if ownerName == "" {
		return nil, fmt.Errorf("error no statefulset found as the contoller owner for the %s pod", pod.Name)
	}

	statefulSetDetails, err := client.AppsV1().StatefulSets(namespace).Get(context.TODO(), ownerName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// the StatefulSet get API returns the statefulsets with the apiVersion and kind field as empty, so setting them explicitly
	statefulSetDetails.TypeMeta.APIVersion = utils.ApiVersionAppsV1
	statefulSetDetails.TypeMeta.Kind = utils.KindStatefulSet

	return statefulSetDetails, nil

}

// GetStatefulSetEventsByName retrieves all StatefulSet events in a given namespace that match the provided statefulSet name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - statefulSetName: the statefulSet name to fetch the events for.
//   - namespace: the namespace in which to search for the StatefulSets.
func GetStatefulSetEventsByName(client kubernetes.Interface, statefulSetName, namespace string) (map[string][]corev1.Event, error) {

	statefulSetEventsMap := make(map[string][]corev1.Event)

	fieldSelector := fmt.Sprintf("involvedObject.name=%s", statefulSetName)

	statefulSetEvents, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	statefulSetEventsMap[statefulSetName] = statefulSetEvents.Items

	return statefulSetEventsMap, nil

}
