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

	// the pod list API returns the pods with the apiVersion and kind field as empty, so setting them explicitly
	for index := range podList.Items {
		podList.Items[index].TypeMeta.APIVersion = utils.ApiVersionV1
		podList.Items[index].TypeMeta.Kind = utils.KindPod
	}

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

	podLogsMap := make(map[string]utils.PodLogs)

	for _, pod := range pods {

		hasContainerRestarted := false
		hasLastTerminationStateTerminated := false // for cases when the restartCount>0 but there are not previous logs

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

	podDescribeMap := make(map[string]string)

	for _, pod := range pods {

		describeSettings := describe.DescriberSettings{
			ShowEvents: includeEvents,
		}

		restClientGetter := genericclioptions.NewConfigFlags(true)

		resourceDescriber, err := describe.Describer(restClientGetter, &meta.RESTMapping{
			GroupVersionKind: corev1.SchemeGroupVersion.WithKind(utils.KindPod),
		})
		if err != nil {
			return nil, err
		}

		podDescribeData, err := resourceDescriber.Describe(namespace, pod.Name, describeSettings)
		if err != nil {
			return podDescribeMap, err
		}

		podDescribeMap[pod.Name] = podDescribeData

	}

	return podDescribeMap, nil

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

	podEventsMap := make(map[string][]corev1.Event)

	for _, pod := range pods {

		fieldSelector := fmt.Sprintf("involvedObject.name=%s", pod.Name)

		eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return podEventsMap, err
		}

		podEventsMap[pod.Name] = eventList.Items

	}

	return podEventsMap, nil

}
