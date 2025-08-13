package controllerrevisions

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetControllerRevisionsBySelector retrieves all Controller revisions in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the Controller Revisions.
//   - namespace: the namespace in which to search for the Controller Revisions.
func GetControllerRevisionsBySelector(client kubernetes.Interface, selector, namespace string) ([]appsv1.ControllerRevision, error) {

	controllerRevisionList, err := client.AppsV1().ControllerRevisions(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the ControllerRevision list API returns the controller-revisions with the apiVersion and kind field as empty, so setting them explicitly
	for index := range controllerRevisionList.Items {
		controllerRevisionList.Items[index].TypeMeta.APIVersion = utils.ApiVersionAppsV1
		controllerRevisionList.Items[index].TypeMeta.Kind = utils.KindControllerRevision
	}

	return controllerRevisionList.Items, err

}
