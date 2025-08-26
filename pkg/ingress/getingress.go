package ingress

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListAllIngressInNamespace list all ingresses in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - namespace: the namespace in which to search for the ingresses.
func ListAllIngressInNamespace(client kubernetes.Interface, namespace string) ([]networkingv1.Ingress, error) {

	ingressList, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// the ingress list API returns the ingresses with the apiVersion and kind field as empty, so setting them explicitly
	for index := range ingressList.Items {
		ingressList.Items[index].TypeMeta.APIVersion = utils.ApiVersionNetworkingV1
		ingressList.Items[index].TypeMeta.Kind = utils.KindIngress
	}

	return ingressList.Items, nil

}
