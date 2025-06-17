package service

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetServiceDetailsBySelector retrieves all services in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the services.
//   - namespace: the namespace in which to search for the services.
func GetServiceDetailsBySelector(client kubernetes.Interface, selector, namespace string) ([]corev1.Service, error) {

	serviceList, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the service list API returns the services with the apiVersion and kind field as empty, so setting them explicitly
	for index := range serviceList.Items {
		serviceList.Items[index].TypeMeta.APIVersion = utils.ApiVersionV1
		serviceList.Items[index].TypeMeta.Kind = utils.KindService
	}

	return serviceList.Items, err

}
