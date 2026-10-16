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

package rbac

import (
	"context"
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// GetRoleBindings retrieves all the RoleBindings in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - namespace: the namespace in which to search for the RoleBindings.
func GetRoleBindings(client kubernetes.Interface, namespace string) ([]rbacv1.RoleBinding, error) {
	return getRoleBindingDetails(client, namespace, metav1.ListOptions{})
}

// GetRoleBindingsBySelector retrieves all the RoleBindings in a given namespace that match the provided label selector.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the RoleBindings.
//   - namespace: the namespace in which to search for the RoleBindings.
func GetRoleBindingsBySelector(client kubernetes.Interface, selector, namespace string) ([]rbacv1.RoleBinding, error) {
	return getRoleBindingDetails(client, namespace, metav1.ListOptions{LabelSelector: selector})
}

func getRoleBindingDetails(client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]rbacv1.RoleBinding, error) {

	roleBindingsList, err := client.RbacV1().RoleBindings(namespace).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("listing Role Binding's in namespace %s: %w", namespace, err)
	}

	// the role-binding list API returns the pods with the apiVersion and kind field as empty, so setting them explicitly
	for i := range roleBindingsList.Items {
		roleBindingsList.Items[i].TypeMeta.APIVersion = utils.ApiVersionRbacV1
		roleBindingsList.Items[i].TypeMeta.Kind = utils.KindRoleBinding
	}

	return roleBindingsList.Items, nil
}

// GetRoleDetailsByName retrieves the Role details by role name in a given namespace.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - name: the name of the Role.
//   - namespace: the namespace in which to search for the Role.
func GetRoleDetailsByName(client kubernetes.Interface, name, namespace string) (rbacv1.Role, error) {

	role, err := client.RbacV1().Roles(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	role.TypeMeta.APIVersion = utils.ApiVersionRbacV1
	role.TypeMeta.Kind = utils.KindRole

	return *role, err

}

// GetClusterRoleDetailsByName retrieves the ClusterRole details by cluster role name.
// ClusterRoles are cluster scoped, so no namespace is required.
// Parameters:
//   - client: the Kubernetes client used to interact with the cluster.
//   - name: the name of the ClusterRole.
func GetClusterRoleDetailsByName(client kubernetes.Interface, name string) (rbacv1.ClusterRole, error) {

	clusterRole, err := client.RbacV1().ClusterRoles().Get(context.TODO(), name, metav1.GetOptions{})

	clusterRole.TypeMeta.APIVersion = utils.ApiVersionRbacV1
	clusterRole.TypeMeta.Kind = utils.KindClusterRole

	return *clusterRole, err
}
