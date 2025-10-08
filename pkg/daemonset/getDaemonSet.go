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

package daemonset

import (
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetDaemonSetByPodName retrieves all daemonsets in a given namespace that match the provided pod name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - podName: the pod name used to filter the daemonsets.
//   - namespace: the namespace in which to search for the daemonsets.
func GetDaemonSetByPodName(client kubernetes.Interface, podName, namespace string) (*appsv1.DaemonSet, error) {

	pod, err := pods.GetPodByName(client, podName, namespace)
	if err != nil {
		return nil, err
	}

	daemonSetName := ""
	for _, owner := range pod.ObjectMeta.OwnerReferences {
		if *owner.Controller {
			daemonSetName = owner.Name
			break
		}
	}
	if daemonSetName == "" {
		return nil, fmt.Errorf("error no daemonset found as the contoller owner for the %s pod", podName)
	}

	daemonSetDetails, err := client.AppsV1().DaemonSets(namespace).Get(context.TODO(), daemonSetName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// the DaemonSet get API returns the daemon-sets with the apiVersion and kind field as empty, so setting them explicitly
	daemonSetDetails.TypeMeta.APIVersion = utils.ApiVersionAppsV1
	daemonSetDetails.TypeMeta.Kind = utils.KindDaemonSet

	return daemonSetDetails, nil
}

// GetDaemonSetEventsByName retrieves all daemonset events in a given namespace that match the provided daemonset name.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - daemonSetName: the daemonset name to fetch the events for.
//   - namespace: the namespace in which to search for the daemonsets.
func GetDaemonSetEventsByName(client kubernetes.Interface, daemonSetName, namespace string) (map[string][]corev1.Event, error) {

	daemonSetEventsMap := make(map[string][]corev1.Event)

	fieldSelector := fmt.Sprintf("involvedObject.name=%s", daemonSetName)

	eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	daemonSetEventsMap[daemonSetName] = eventList.Items

	return daemonSetEventsMap, nil

}
