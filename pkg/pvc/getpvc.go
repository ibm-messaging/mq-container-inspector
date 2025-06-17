package pvc

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPVCDetailsBySelector retrieves all pvc's in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pvc's.
//   - namespace: the namespace in which to search for the pvc's.
func GetPVCDetailsBySelector(client kubernetes.Interface, selector, namespace string) ([]corev1.PersistentVolumeClaim, error) {

	pvcList, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the pvc list API returns the pvc's with the apiVersion and kind field as empty, so setting them explicitly
	for index := range pvcList.Items {
		pvcList.Items[index].TypeMeta.APIVersion = utils.ApiVersionV1
		pvcList.Items[index].TypeMeta.Kind = utils.KindPVC
	}

	return pvcList.Items, err

}
