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
package routes

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"sigs.k8s.io/yaml"

	routeV1 "github.com/openshift/api/route/v1"
)

// WriteRouteDetailsBySelectorToFile writes the Routes details to a separate YAML file.
// Parameters:
//   - routeList:        the list of routes whose details will be written.
//   - fileNameFormat:   the format string used to name each file; must contain one "%s", which will be replaced by the route name.
//   - outputDir:        the directory in which the YAML files will be created.
func WriteRouteDetailsBySelectorToFile(routeList []routeV1.Route, fileNameFormat, outputDir string) error {

	for _, route := range routeList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, route.Name)

		data, err := yaml.Marshal(route)
		if err != nil {
			return fmt.Errorf("error marshalling yaml for route %s: %v", route.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0o600); err != nil {
			return fmt.Errorf("error while writing route %s data to yaml file %s: %v", route.Name, fileName, err)
		}

	}

	return nil

}
