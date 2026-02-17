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

package serviceaccount

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetServiceAccountDetailsBySelector retrieves all service-accounts in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the service-accounts.
//   - namespace: the namespace in which to search for the service-accounts.
func GetServiceAccountDetailsBySelector(client kubernetes.Interface, selector, namespace string) ([]corev1.ServiceAccount, error) {

	serviceAccountList, err := client.CoreV1().ServiceAccounts(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the service-account list API returns the service accounts with the apiVersion and kind field as empty, so setting them explicitly
	for index := range serviceAccountList.Items {
		serviceAccountList.Items[index].TypeMeta.APIVersion = utils.ApiVersionV1
		serviceAccountList.Items[index].TypeMeta.Kind = utils.KindServiceAccount
	}

	return serviceAccountList.Items, err
}

