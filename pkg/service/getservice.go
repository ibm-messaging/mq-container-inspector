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
package service

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetServiceDetailsByPodName retrieves all services in a given namespace that match the provided pod name
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - podName: the pod name used to filter the services.
//   - namespace: the namespace in which to search for the services.
func GetServiceDetailsByPodName(client kubernetes.Interface, podName, namespace string) ([]corev1.Service, error) {

	podList, err := pods.GetMQReplicaPodsViaService(client, podName, namespace)
	if err != nil {
		return nil, err
	}

	var podLabels []map[string]string

	pod := podList[0]
	podLabels = append(podLabels, pod.ObjectMeta.Labels)

	// if the pod instance is of type NativeHA then we will be having additional services for communication between pods
	if utils.GetPodInstance(&pod) == utils.NativeHA {
		for index, pod := range podList {
			if index < 1 {
				continue
			}
			podLabels = append(podLabels, pod.ObjectMeta.Labels)
		}
	}

	// fetch all the services in the namespace
	serviceList, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var matchingServices []corev1.Service
	for _, podLabel := range podLabels {
		for _, service := range serviceList.Items {

			// check if the service selector is a subset of the pod label
			if utils.IsSelectorSubsetOfLabels(service.Spec.Selector, podLabel) {
				matchingServices = append(matchingServices, service)
			}

		}
	}

	return matchingServices, nil

}
