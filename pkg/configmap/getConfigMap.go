/*
© Copyright IBM Corporation 2025,2026

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
package configmap

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetConfigMapDetailsBySelector gets the configmap details by label selector in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: he label selector used to filter the configmap's
//   - namespace: the namespace in which to search for the configmap.
func GetConfigMapDetailsBySelector(client kubernetes.Interface, selector, namespace string) ([]corev1.ConfigMap, error) {
	configMap, err := client.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the configmap get API returns the configmap with the apiVersion and kind field as empty, so setting them explicitly
	for index := range configMap.Items {
		configMap.Items[index].TypeMeta.APIVersion = utils.ApiVersionV1
		configMap.Items[index].TypeMeta.Kind = utils.KindConfigMap
	}

	return configMap.Items, err
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
