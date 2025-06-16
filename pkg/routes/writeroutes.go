package routes

import (
	"fmt"
	"os"

	routeV1 "github.com/openshift/api/route/v1"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"sigs.k8s.io/yaml"
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

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing route %s data to yaml file %s: %v", route.Name, fileName, err)
		}

	}

	return nil

}
