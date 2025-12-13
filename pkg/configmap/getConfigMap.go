package configmap

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetConfigMapDetailsByName gets the configmap details by name in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - name: the configmap name to fetch the details for.
//   - namespace: the namespace in which to search for the configmap.
func GetConfigMapDetailsByName(client kubernetes.Interface, name, namespace string) (*corev1.ConfigMap, error) {
	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// the configmap get API returns the configmap with the apiVersion and kind field as empty, so setting them explicitly
	configMap.TypeMeta.APIVersion = utils.ApiVersionV1
	configMap.TypeMeta.Kind = utils.KindConfigMap

	return configMap, nil
}

// DeleteConfigMapByName delete's the configmap by name in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - name: the configmap name to delete.
//   - namespace: the namespace in which to search for the configmap.
func DeleteConfigMapByName(client kubernetes.Interface, name, namespace string) error {
	return client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// UpdateConfigMapByName update's the configmap by name in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - name: the configmap name to update.
//   - namespace: the namespace in which to search for the configmap.
func UpdateConfigMapByName(client kubernetes.Interface, configMap *corev1.ConfigMap, namespace string) (*corev1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
}

// CreateConfigMapByName create's the configmap by name in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - name: the configmap name to create.
//   - namespace: the namespace in which to search for the configmap.
func CreateConfigMapByName(client kubernetes.Interface, configMap *corev1.ConfigMap, namespace string) (*corev1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
}
