/*
© Copyright IBM Corporation 2026

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
package replicaset

import (
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetReplicaSetBySelector retrieves all replicasets in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the replicasets.
//   - namespace: the namespace in which to search for the replicasets.
func GetReplicaSetBySelector(client kubernetes.Interface, selector, namespace string) ([]appsv1.ReplicaSet, error) {

	replicaSetList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the ReplicaSet list API returns the replicasets with the apiVersion and kind field as empty, so setting them explicitly
	for index := range replicaSetList.Items {
		replicaSetList.Items[index].TypeMeta.APIVersion = utils.ApiVersionAppsV1
		replicaSetList.Items[index].TypeMeta.Kind = utils.KindReplicaSet
	}

	return replicaSetList.Items, err

}

// GetEventsForReplicaSetList retrieves all events for the provided ReplicaSetList.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - replicasetList: the replicasets for which to collect the events
//   - namespace: the namespace in which to search for the replicasets.
func GetEventsForReplicaSetList(client kubernetes.Interface, replicasetList []appsv1.ReplicaSet) (map[string][]corev1.Event, error) {

	replicasetEventsMap := make(map[string][]corev1.Event)

	for _, replicaset := range replicasetList {

		fieldSelector := fmt.Sprintf("involvedObject.name=%s", replicaset.ObjectMeta.Name)

		eventList, err := client.CoreV1().Events(replicaset.ObjectMeta.Namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return nil, err
		}

		replicasetEventsMap[replicaset.ObjectMeta.Name] = eventList.Items

	}

	return replicasetEventsMap, nil

}
