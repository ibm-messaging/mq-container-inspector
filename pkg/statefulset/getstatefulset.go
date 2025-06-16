package statefulset

import (
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetStatefulSetDetailsBySelector retrieves all StatefulSets in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
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

// GetStatefulSetRevisionsBySelector retrieves all StatefulSet revisions in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
func GetStatefulSetRevisionsBySelector(client kubernetes.Interface, selector, namespace string) ([]appsv1.ControllerRevision, error) {

	statefulSetRevisionList, err := client.AppsV1().ControllerRevisions(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the ControllerRevision list API returns the controller-revisions with the apiVersion and kind field as empty, so setting them explicitly
	for index := range statefulSetRevisionList.Items {
		statefulSetRevisionList.Items[index].TypeMeta.APIVersion = utils.ApiVersionAppsV1
		statefulSetRevisionList.Items[index].TypeMeta.Kind = utils.KindControllerRevision
	}

	return statefulSetRevisionList.Items, err

}

// GetStatefulSetRevisionsBySelector retrieves all StatefulSet events in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
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
