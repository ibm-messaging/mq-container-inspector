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

package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/rbac"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/serviceaccount"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
)

func gatherRBACToFiles(client kubernetes.Interface, flags utils.MustGatherFlags, pod *corev1.Pod, logger *slog.Logger) error {

	rbacDirectory := filepath.Join(flags.OutputDir, "rbac")
	if !utils.CheckIfDirectoryExist(rbacDirectory) {
		if err := utils.CreateDirectory(rbacDirectory, 0o775); err != nil {
			return err
		}
	}

	if pod != nil {
		gatherRBACDetailsByPod(client, pod, flags.QueueManagerNamespace, rbacDirectory, logger)
	} else if flags.QueueManagerName != "" {
		gatherRBACDetailsByOwnerReference(client, flags.QueueManagerName, flags.QueueManagerNamespace, rbacDirectory, logger)
	}

	// count of directories and in each directory the file-count
	utils.LogDirectorySummary(rbacDirectory, logger)

	return nil

}

func gatherRBACDetailsByOwnerReference(client kubernetes.Interface, qmName, namespace, parentDir string, logger *slog.Logger) {

	labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", qmName)

	// Collect the roleBindings
	roleBindings, err := rbac.GetRoleBindingsBySelector(client, labelSelector, namespace)
	if err != nil {
		logger.Info(fmt.Sprintf("error fetching role bindings for %s label selector in %s namespace: %v", labelSelector, namespace, err))
		return
	} else if len(roleBindings) == 0 {
		logger.Info(fmt.Sprintf("no role bindings for %s label selector in %s namespace", labelSelector, namespace))
		return
	}

	// Filter based on the owner
	var filteredRoleBindings []rbacv1.RoleBinding

	for _, roleBinding := range roleBindings {

		var ownerMatched bool

		for _, owner := range roleBinding.ObjectMeta.OwnerReferences {
			if owner.Kind == utils.KindQueueManager && owner.Controller != nil && *owner.Controller && owner.Name == qmName {
				ownerMatched = true
				break
			}
		}

		if ownerMatched {
			filteredRoleBindings = append(filteredRoleBindings, roleBinding)
		}

	}

	if len(filteredRoleBindings) == 0 {
		logger.Info(fmt.Sprintf("no role bindings owned by the %s queueManager in %s namespace", qmName, namespace))
		return
	}

	if err := writeRoleBindingsToFiles(filteredRoleBindings, parentDir, logger); err != nil {
		logger.Info(fmt.Sprintf("error writing %d role bindings to %s directory: %v", len(filteredRoleBindings), parentDir, err))
	}

	// Collect the roles/clusterRoles
	collectRolesAndClusterRoleDetails(client, filteredRoleBindings, parentDir, namespace, logger)

	// Collect the ServiceAccount details referenced by the roleBinding subjects
	collectSubjectServiceAccountDetails(client, filteredRoleBindings, parentDir, namespace, nil, logger)

}

func gatherRBACDetailsByPod(client kubernetes.Interface, pod *corev1.Pod, namespace, parentDir string, logger *slog.Logger) {

	// Get the ServiceAccount name from the pod
	serviceAccountName := pod.Spec.ServiceAccountName

	if serviceAccountName == "" {
		logger.Info(fmt.Sprintf("no service-account attached to the pod: %s", pod.ObjectMeta.Name))
		return
	}

	// Collect the ServiceAccount Details
	serviceAccount, err := serviceaccount.GetServiceAccountDetailsByName(client, serviceAccountName, namespace)
	if err != nil {
		logger.Info(fmt.Sprintf("error fetching %s service-account attached to %s pod: %v", serviceAccountName, pod.ObjectMeta.Name, err))
		return
	}

	if err := writeServiceAccountToFiles([]corev1.ServiceAccount{*serviceAccount}, parentDir, logger); err != nil {
		logger.Info(fmt.Sprintf("error writing 1 service account to %s directory: %v", parentDir, err))
	}

	// Collect the RoleBindings
	roleBindings, err := rbac.GetRoleBindings(client, namespace)
	if err != nil {
		logger.Info(fmt.Sprintf("error fetching role bindings in %s namespace: %v", namespace, err))
		return
	} else if len(roleBindings) == 0 {
		logger.Info(fmt.Sprintf("no role bindings in %s namespace", namespace))
		return
	}

	var filteredRoleBindings []rbacv1.RoleBinding

	for _, roleBinding := range roleBindings {

		var serviceAccountSubjectMatched bool
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == utils.KindServiceAccount && subject.Name == serviceAccountName && subjectNamespace(subject, roleBinding.Namespace) == pod.ObjectMeta.Namespace {
				serviceAccountSubjectMatched = true
				break
			}
		}

		if serviceAccountSubjectMatched {
			filteredRoleBindings = append(filteredRoleBindings, roleBinding)
		}

	}

	if len(filteredRoleBindings) == 0 {
		logger.Info(fmt.Sprintf("no role bindings referencing the %s service-account in %s namespace", serviceAccountName, namespace))
		return
	}

	if err := writeRoleBindingsToFiles(filteredRoleBindings, parentDir, logger); err != nil {
		logger.Info(fmt.Sprintf("error writing %d role bindings to %s directory: %v", len(filteredRoleBindings), parentDir, err))
	}

	// Collect the roles/ClusterRoles
	collectRolesAndClusterRoleDetails(client, filteredRoleBindings, parentDir, namespace, logger)

	// Collect any other ServiceAccounts referenced by RoleBinding subjects
	// The pod's own service-account has been collected above
	collectSubjectServiceAccountDetails(client, filteredRoleBindings, parentDir, namespace, map[string]struct{}{serviceAccountName: {}}, logger)

}

func collectRolesAndClusterRoleDetails(client kubernetes.Interface, roleBindings []rbacv1.RoleBinding, parentDir, namespace string, logger *slog.Logger) {
	var roleNames []string
	var clusterRoleNames []string

	seenRoles := make(map[string]struct{})
	seenClusterRoles := make(map[string]struct{})

	for _, roleBinding := range roleBindings {

		name := roleBinding.RoleRef.Name

		switch roleBinding.RoleRef.Kind {
		case utils.KindRole:
			if _, ok := seenRoles[name]; ok {
				continue
			}
			seenRoles[name] = struct{}{}
			roleNames = append(roleNames, name)
		case utils.KindClusterRole:
			if _, ok := seenClusterRoles[name]; ok {
				continue
			}
			seenClusterRoles[name] = struct{}{}
			clusterRoleNames = append(clusterRoleNames, name)
		}
	}

	var roles []rbacv1.Role

	for _, name := range roleNames {
		role, err := rbac.GetRoleDetailsByName(client, name, namespace)
		if err != nil {
			logger.Info(fmt.Sprintf("error fetching %s role: %v", name, err))
			continue
		}
		roles = append(roles, role)
	}

	var clusterRoles []rbacv1.ClusterRole

	for _, name := range clusterRoleNames {
		clusterRole, err := rbac.GetClusterRoleDetailsByName(client, name)
		if err != nil {
			logger.Info(fmt.Sprintf("error fetching %s cluster role: %v", name, err))
			continue
		}
		clusterRoles = append(clusterRoles, clusterRole)
	}

	if err := writeRolesToFiles(roles, parentDir, logger); err != nil {
		logger.Info(fmt.Sprintf("error writing %d roles to %s directory: %v", len(roles), parentDir, err))
	}

	if err := writeClusterRolesToFiles(clusterRoles, parentDir); err != nil {
		logger.Info(fmt.Sprintf("error writing %d cluster roles to %s directory: %v", len(clusterRoles), parentDir, err))
	}

}

// collectSubjectServiceAccountDetails collects the ServiceAccounts referenced by the
// subjects of the supplied roleBindings.
func collectSubjectServiceAccountDetails(client kubernetes.Interface, roleBindings []rbacv1.RoleBinding, parentDir, namespace string, alreadyCollected map[string]struct{}, logger *slog.Logger) {

	seen := make(map[string]struct{})
	for name := range alreadyCollected {
		seen[name] = struct{}{}
	}

	var serviceAccountNames []string

	for _, roleBinding := range roleBindings {

		for _, subject := range roleBinding.Subjects {

			if subject.Kind != utils.KindServiceAccount {
				continue
			}

			// the subject namespace defaults to the roleBinding namespace when not set
			if subjectNamespace(subject, roleBinding.Namespace) != namespace {
				continue
			}

			if _, ok := seen[subject.Name]; ok {
				continue
			}

			seen[subject.Name] = struct{}{}
			serviceAccountNames = append(serviceAccountNames, subject.Name)

		}

	}

	var serviceAccounts []corev1.ServiceAccount

	for _, name := range serviceAccountNames {
		serviceAccount, err := serviceaccount.GetServiceAccountDetailsByName(client, name, namespace)
		if err != nil {
			logger.Info(fmt.Sprintf("error fetching %s service-account: %v", name, err))
			continue
		}
		serviceAccounts = append(serviceAccounts, *serviceAccount)
	}

	if err := writeServiceAccountToFiles(serviceAccounts, parentDir, logger); err != nil {
		logger.Info(fmt.Sprintf("error writing %d service account to %s directory: %v", len(serviceAccounts), parentDir, err))
	}
}

// subjectNamespace returns the namespace of a roleBinding subject, defaulting to the
// namespace of the roleBinding itself when the subject does not specify one.
func subjectNamespace(subject rbacv1.Subject, roleBindingNamespace string) string {
	if subject.Namespace != "" {
		return subject.Namespace
	}
	return roleBindingNamespace
}

func writeRoleBindingsToFiles(roleBindings []rbacv1.RoleBinding, parentDir string, logger *slog.Logger) error {

	roleBindingDir := filepath.Join(parentDir, "rolebindings")
	if !utils.CheckIfDirectoryExist(roleBindingDir) {
		if err := utils.CreateDirectory(roleBindingDir, 0o775); err != nil {
			return err
		}
	}

	fileNameFormat := "%s.yaml"
	if err := rbac.WriteRoleBindingYamlsToFile(roleBindings, fileNameFormat, roleBindingDir); err != nil {
		return err
	}

	if fileCount, err := utils.GetFileCountInDirectory(roleBindingDir); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("RoleBinding details: %s: Total Files: %d", roleBindingDir, fileCount))
	}

	return nil

}

func writeServiceAccountToFiles(serviceAccounts []corev1.ServiceAccount, parentDir string, logger *slog.Logger) error {

	serviceAccountDir := filepath.Join(parentDir, "serviceaccounts")
	if !utils.CheckIfDirectoryExist(serviceAccountDir) {
		if err := utils.CreateDirectory(serviceAccountDir, 0o775); err != nil {
			return err
		}
	}

	fileNameFormat := "%s.yaml"
	if err := serviceaccount.WriteServiceAccountYamlsToFile(serviceAccounts, fileNameFormat, serviceAccountDir); err != nil {
		return err
	}

	if fileCount, err := utils.GetFileCountInDirectory(serviceAccountDir); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("ServiceAccount details: %s: Total Files: %d", serviceAccountDir, fileCount))
	}

	return nil

}

func writeRolesToFiles(roles []rbacv1.Role, parentDir string, logger *slog.Logger) error {

	roleDir := filepath.Join(parentDir, "roles")
	if !utils.CheckIfDirectoryExist(roleDir) {
		if err := utils.CreateDirectory(roleDir, 0o775); err != nil {
			return err
		}
	}

	fileNameFormat := "%s.yaml"
	if err := rbac.WriteRoleYamlsToFile(roles, fileNameFormat, roleDir); err != nil {
		return err
	}

	if fileCount, err := utils.GetFileCountInDirectory(roleDir); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Role details: %s: Total Files: %d", roleDir, fileCount))
	}

	return nil

}

func writeClusterRolesToFiles(clusterRoles []rbacv1.ClusterRole, parentDir string) error {

	clusterRoleDir := filepath.Join(parentDir, "clusterroles")
	if !utils.CheckIfDirectoryExist(clusterRoleDir) {
		if err := utils.CreateDirectory(clusterRoleDir, 0o775); err != nil {
			return err
		}
	}

	fileNameFormat := "%s.yaml"
	if err := rbac.WriteClusterRoleYamlsToFile(clusterRoles, fileNameFormat, clusterRoleDir); err != nil {
		return err
	}

	return nil

}
