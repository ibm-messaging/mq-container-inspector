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

package ingress

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
