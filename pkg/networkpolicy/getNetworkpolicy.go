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

package networkpolicy

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	networkV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetNetworkPolicyBySelector retrieves all network-policies in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the network-policies.
//   - namespace: the namespace in which to search for the network-policies.
func GetNetworkPolicyBySelector(client kubernetes.Interface, selector, namespace string) ([]networkV1.NetworkPolicy, error) {

	networkPolicyList, err := client.NetworkingV1().NetworkPolicies(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the NetworkPolicy list API returns the network-policies with the apiVersion and kind field as empty, so setting them explicitly
	for index := range networkPolicyList.Items {
		networkPolicyList.Items[index].TypeMeta.APIVersion = utils.ApiVersionNetworkingV1
		networkPolicyList.Items[index].TypeMeta.Kind = utils.KindNetworkPolicy
	}

	return networkPolicyList.Items, err

}
