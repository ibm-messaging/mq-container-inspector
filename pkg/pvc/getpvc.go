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
package pvc

import (
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
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

// GetPVCDetailByName retrieves the pvc in a given namespace that match the provided pvc name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - pvcName: the pvc name.
//   - namespace: the namespace in which to search for the pvc.
func GetPVCDetailByName(client kubernetes.Interface, pvcName, namespace string) (*corev1.PersistentVolumeClaim, error) {

	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// the pvc get API returns the pvc's with the apiVersion and kind field as empty, so setting them explicitly
	pvc.TypeMeta.APIVersion = utils.ApiVersionV1
	pvc.TypeMeta.Kind = utils.KindPVC

	return pvc, nil

}

// GetPVCDetailsByPodName retrieves all pvc's in a given namespace that match the provided pod name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - podName: the pod name used to filter the pvc's.
//   - namespace: the namespace in which to search for the pvc's.
func GetPVCDetailsByPodName(client kubernetes.Interface, podName, namespace string) ([]corev1.PersistentVolumeClaim, error) {

	// fetch all the pods
	pods, err := pods.GetMQReplicaPodsViaService(client, podName, namespace)
	if err != nil {
		return nil, err
	}

	var podAttachedVolumeNames []string

	for _, pod := range pods {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				podAttachedVolumeNames = append(podAttachedVolumeNames, volume.PersistentVolumeClaim.ClaimName)
			}
		}
	}

	if len(podAttachedVolumeNames) == 0 {
		return nil, fmt.Errorf("no attached pvc's found for %s pod", podName)
	}

	var pvcList []corev1.PersistentVolumeClaim

	for _, pvcName := range podAttachedVolumeNames {

		pvc, err := GetPVCDetailByName(client, pvcName, namespace)
		if err != nil {
			return nil, err
		}

		pvcList = append(pvcList, *pvc)

	}

	return pvcList, nil

}
