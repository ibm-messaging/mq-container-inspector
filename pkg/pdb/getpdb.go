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

package pdb

import (
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPdbDetails retrieves all the PodDisruptionBudgets in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - namespace: the namespace in which to search for the PodDisruptionBudgets.
func GetPdbDetails(client kubernetes.Interface, namespace string) ([]policyv1.PodDisruptionBudget, error) {
	return listPdbs(client, namespace, metav1.ListOptions{})
}

// GetPdbDetailsBySelector retrieves all the PodDisruptionBudgets in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the PodDisruptionBudgets.
//   - namespace: the namespace in which to search for the PodDisruptionBudgets.
func GetPdbDetailsBySelector(client kubernetes.Interface, selector, namespace string) ([]policyv1.PodDisruptionBudget, error) {
	return listPdbs(client, namespace, metav1.ListOptions{LabelSelector: selector})
}

func listPdbs(client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]policyv1.PodDisruptionBudget, error) {

	pdbList, err := client.PolicyV1().PodDisruptionBudgets(namespace).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("listing PDBs in namespace %s: %w", namespace, err)
	}

	// the pdb list API returns the pods with the apiVersion and kind field as empty, so setting them explicitly
	for i := range pdbList.Items {
		pdbList.Items[i].TypeMeta.APIVersion = utils.ApiVersionPolicyV1
		pdbList.Items[i].TypeMeta.Kind = utils.KindPodDisruptionBudget
	}

	return pdbList.Items, nil
}
