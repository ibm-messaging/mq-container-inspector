package deployment

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
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
		deploymentList.Items[index].TypeMeta.APIVersion = utils.ApiVersionV1
		deploymentList.Items[index].TypeMeta.Kind = utils.KindDeployment
	}

	return deploymentList.Items, err

}
