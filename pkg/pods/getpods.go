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
package pods

import (
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	describe "k8s.io/kubectl/pkg/describe"
)

// GetPodBySelector retrieves all pods in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
func GetPodsBySelector(client kubernetes.Interface, selector, namespace string) ([]corev1.Pod, error) {

	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	podList = EditPodDetails(podList)

	return podList.Items, err

}

// GetPodLogsBySelector retrieves all the pods previous and current logs in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
//   - containerName: the container for which we want the logs for
func GetPodLogsBySelector(client kubernetes.Interface, selector, namespace, containerName string) (map[string]utils.PodLogs, error) {

	pods, err := GetPodsBySelector(client, selector, namespace)
	if err != nil {
		return nil, err
	}

	podLogsMap := GetPodLogs(client, pods, namespace, containerName)

	return podLogsMap, err

}

// GetPodDescribeBySelector retrieves all the pods describe logs in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
//   - includeEvents: whether to include pod events in the describe output
func GetPodDescribeBySelector(client kubernetes.Interface, selector, namespace string, includeEvents bool) (map[string]string, error) {

	pods, err := GetPodsBySelector(client, selector, namespace)
	if err != nil {
		return nil, err
	}

	podDescribeMap, err := GetPodDescribeLogs(client, pods, namespace)

	return podDescribeMap, err

}

// GetPodEventsBySelector retrieves all the pods events in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
func GetPodEventsBySelector(client kubernetes.Interface, selector, namespace string) (map[string][]corev1.Event, error) {

	pods, err := GetPodsBySelector(client, selector, namespace)
	if err != nil {
		return nil, err
	}

	podEventsMap, err := GetPodEvents(client, pods, namespace)

	return podEventsMap, err

}

// GetPodByName retrieves the pod details by pod name in a given namespace.
//   - client: the Kubernetes client used to interact with the cluster.
//   - podName: the name of the pod.
//   - namespace: the namespace in which to search for the pods.
func GetPodByName(client kubernetes.Interface, podName, namespace string) (*corev1.Pod, error) {
	return client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})

}

// GetMQReplicaPodsViaService retrieves the pod list by pod name in a given namespace.
//   - client: the Kubernetes client used to interact with the cluster.
//   - podName: the name of the pod.
//   - namespace: the namespace in which to search for the pods.
func GetMQReplicaPodsViaService(client kubernetes.Interface, podName, namespace string) ([]corev1.Pod, error) {

	pod, err := GetPodByName(client, podName, namespace)
	if err != nil {
		return nil, err
	}

	// get the expected pod-count
	expectedPodCount := 1
	switch utils.GetPodInstance(pod) {
	case utils.NativeHA:
		expectedPodCount = 3
	case utils.MultiInstance:
		expectedPodCount = 2
	}

	if expectedPodCount == 1 {
		// the pod get API returns the pods with the apiVersion and kind field as empty, so setting them explicitly
		pod.TypeMeta.APIVersion = utils.ApiVersionV1
		pod.TypeMeta.Kind = utils.KindPod
		return []corev1.Pod{*pod}, nil
	}

	// fetch all the services in the namespace
	serviceList, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, service := range serviceList.Items {

		serviceSelector := service.Spec.Selector
		if len(serviceSelector) == 0 {
			continue
		}

		selectorString := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: serviceSelector})

		matchingPods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selectorString,
		})
		if err != nil {
			return nil, err
		}

		isTargetPodIncluded := false
		for _, pod := range matchingPods.Items {
			if pod.Name == podName {
				isTargetPodIncluded = true
				break
			}
		}

		if isTargetPodIncluded && len(matchingPods.Items) == expectedPodCount {
			podList := EditPodDetails(matchingPods)
			return podList.Items, nil
		}

	}

	return nil, fmt.Errorf("error could not find a matching service for %s pod in the %s namespace", podName, namespace)

}

// EditPodDetails edit the pod api version and kind.
//   - podList: the list of pods
func EditPodDetails(podList *corev1.PodList) *corev1.PodList {

	// the pod list API returns the pods with the apiVersion and kind field as empty, so setting them explicitly
	for index := range podList.Items {
		podList.Items[index].TypeMeta.APIVersion = utils.ApiVersionV1
		podList.Items[index].TypeMeta.Kind = utils.KindPod
	}

	return podList

}

// GetPodLogs retrieves all the pods previous and current logs in a given namespace
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - pods: the pod list.
//   - namespace: the namespace in which to search for the pods.
//   - containerName: the container for which we want the logs for
func GetPodLogs(client kubernetes.Interface, pods []corev1.Pod, namespace, containerName string) map[string]utils.PodLogs {

	podLogsMap := make(map[string]utils.PodLogs)

	for _, pod := range pods {

		hasContainerRestarted := false
		hasLastTerminationStateTerminated := false // for cases when the restartCount>0 but there are no previous logs

		for _, container := range pod.Status.ContainerStatuses {
			if container.Name == containerName {
				hasContainerRestarted = container.RestartCount > 0
				hasLastTerminationStateTerminated = container.LastTerminationState.Terminated != nil
				break
			}
		}

		var prevPodLogsRequest *rest.Request

		if hasContainerRestarted && hasLastTerminationStateTerminated {
			prevPodLogsRequest = client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
				Container: containerName,
				Previous:  true,
			})
		}

		currentPodLogsRequest := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: containerName,
		})

		podLogs := utils.PodLogs{
			PreviousPodLogsRequest: prevPodLogsRequest,
			CurrentPodLogsRequest:  currentPodLogsRequest,
		}

		podLogsMap[pod.Name] = podLogs
	}

	return podLogsMap

}

// GetPodDescribeLogs retrieves all the pods describe logs in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - pods: the pod list.
//   - namespace: the namespace in which to search for the pods.
//   - includeEvents: whether to include pod events in the describe output
func GetPodDescribeLogs(client kubernetes.Interface, pods []corev1.Pod, namespace string) (map[string]string, error) {

	podDescribeLogsMap := make(map[string]string)

	for _, pod := range pods {

		describerSettings := describe.DescriberSettings{
			ShowEvents: true,
		}

		restClientGetter := genericclioptions.NewConfigFlags(true)

		resourceDescriber, err := describe.Describer(restClientGetter, &meta.RESTMapping{
			GroupVersionKind: corev1.SchemeGroupVersion.WithKind(utils.KindPod),
		})
		if err != nil {
			return nil, err
		}

		podDescribeData, err := resourceDescriber.Describe(namespace, pod.Name, describerSettings)
		if err != nil {
			return nil, err
		}

		podDescribeLogsMap[pod.Name] = podDescribeData

	}

	return podDescribeLogsMap, nil

}

// GetPodEvents retrieves all the pods events in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - pods: the list of pods.
//   - namespace: the namespace in which to search for the pods.
func GetPodEvents(client kubernetes.Interface, pods []corev1.Pod, namespace string) (map[string][]corev1.Event, error) {

	podEventsMap := make(map[string][]corev1.Event)

	for _, pod := range pods {

		fieldSelector := fmt.Sprintf("involvedObject.name=%s", pod.Name)

		eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return nil, err
		}

		podEventsMap[pod.Name] = eventList.Items

	}

	return podEventsMap, nil

}
