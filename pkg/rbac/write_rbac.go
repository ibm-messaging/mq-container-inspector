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
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

// WriteRoleBindingYamlsToFile writes each RoleBinding's details to a separate YAML file.
// Parameters:
//   - roleBindings:   the list of RoleBindings whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the RoleBinding name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteRoleBindingYamlsToFile(roleBindings []rbacv1.RoleBinding, fileNameFormat, outputDir string) error {

	for _, roleBinding := range roleBindings {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, roleBinding.ObjectMeta.Name)

		data, err := yaml.Marshal(roleBinding)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for role-binding %s: %v", roleBinding.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return fmt.Errorf("error while writing role-binding %s data in the yaml file: %v", roleBinding.ObjectMeta.Name, err)
		}

	}

	return nil

}

// WriteRoleYamlsToFile writes each Role's details to a separate YAML file.
// Parameters:
//   - roles:          the list of Roles whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the Role name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteRoleYamlsToFile(roles []rbacv1.Role, fileNameFormat, outputDir string) error {

	for _, role := range roles {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, role.ObjectMeta.Name)

		data, err := yaml.Marshal(role)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for role %s: %v", role.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return fmt.Errorf("error while writing role %s data in the yaml file: %v", role.ObjectMeta.Name, err)
		}

	}

	return nil

}

// WriteClusterRoleYamlsToFile writes each ClusterRole's details to a separate YAML file.
// Parameters:
//   - clusterRoles:   the list of ClusterRoles whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the ClusterRole name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteClusterRoleYamlsToFile(clusterRoles []rbacv1.ClusterRole, fileNameFormat, outputDir string) error {

	for _, clusterRole := range clusterRoles {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, clusterRole.ObjectMeta.Name)

		data, err := yaml.Marshal(clusterRole)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for cluster-role %s: %v", clusterRole.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return fmt.Errorf("error while writing cluster-role %s data in the yaml file: %v", clusterRole.ObjectMeta.Name, err)
		}
	}

	return nil

}
