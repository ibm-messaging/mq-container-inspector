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
package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	routeV1 "github.com/openshift/api/route/v1"
	routeClient "github.com/openshift/client-go/route/clientset/versioned"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/routes"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/service"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func gatherRoutesToFiles(cfg *rest.Config, coreClient kubernetes.Interface, routeClient routeClient.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	var routeListBySelector []routeV1.Route
	var routeListByFieldSelector []routeV1.Route
	var err error

	// since routes are OCP specific, check if routes exist on the cluster
	isRoutePresent, err := checkIfRoutesArePresentInCluster(cfg)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if !isRoutePresent {
		logger.Info(fmt.Sprintf("routes %v are not present in this cluster", utils.RouteAPIGroupName))
		return nil
	}

	logger.Info(fmt.Sprintf("routes %v are present in this cluster", utils.RouteAPIGroupName))

	// create routes directory to store routes files
	routesDirectory := filepath.Join(flags.OutputDir, "routes")
	directoryExist := utils.CheckIfDirectoryExist(routesDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(routesDirectory, 0775); err != nil {
			return err
		}
	}

	if flags.QueueManagerName != "" {
		routeLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)
		routeFieldSelector := fmt.Sprintf("spec.to.name=%s-ibm-mq", flags.QueueManagerName)

		// get routes created by mq-operator
		routeListBySelector, err = routes.GetRouteDetailsBySelector(routeClient, routeLabelSelector, flags.QueueManagerNamespace)
		if err != nil {
			logger.Error(fmt.Sprintf("error while fetching routes with selector %s: %v", routeLabelSelector, err))
		}

		// get routes to the QueueManager
		routeListByFieldSelector, err = routes.GetRouteDetailsByFieldSelector(routeClient, routeFieldSelector, flags.QueueManagerNamespace)
		if err != nil {
			logger.Error(fmt.Sprintf("error while fetching routes with field selector %s: %v", routeFieldSelector, err))
		}
	} else if flags.PodName != "" {

		// find all the services for the pods
		serviceList, err := service.GetServiceDetailsByPodName(coreClient, flags.PodName, flags.QueueManagerNamespace)
		if err != nil {
			logger.Error(fmt.Sprintf("error fetching services for %s pod: %v", flags.PodName, err))
			return nil
		}

		// for every service find the routes pointing to it
		for _, service := range serviceList {
			fieldSelector := fmt.Sprintf("spec.to.name=%s", service.ObjectMeta.Name)

			routeList, err := routes.GetRouteDetailsByFieldSelector(routeClient, fieldSelector, flags.QueueManagerNamespace)
			if err != nil {
				logger.Error("error fetching routes for %s service: %v", service.ObjectMeta.Name, err)
				continue
			}

			routeListByFieldSelector = append(routeListByFieldSelector, routeList...)
		}

	}

	if routeListBySelector != nil {
		// write the route details to their respective yamls
		routesDetailsBySelectorFileNameFormat := "%s-routes.yaml"
		if err := routes.WriteRouteDetailsBySelectorToFile(routeListBySelector, routesDetailsBySelectorFileNameFormat, routesDirectory); err != nil {
			logger.Error(err.Error())
		}
	}

	if routeListByFieldSelector != nil {
		// write the route details to their respective yamls
		routesDetailsByFieldSelectorFileNameFormat := "%s-routes-to-qm.yaml"
		if err := routes.WriteRouteDetailsBySelectorToFile(routeListByFieldSelector, routesDetailsByFieldSelectorFileNameFormat, routesDirectory); err != nil {
			logger.Error(err.Error())
		}
	}

	if fileCount, err := utils.GetFileCountInDirectory(routesDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Route details: %s: Total Files: %d", routesDirectory, fileCount))
	}

	return nil

}

func checkIfRoutesArePresentInCluster(cfg *rest.Config) (bool, error) {

	discoveryClient, err := kubeclient.BuildDiscoveryClientFromConfig(cfg)
	if err != nil {
		return false, fmt.Errorf("error building discover client: %v", err)
	}

	apiGroups, err := discoveryClient.ServerGroups()
	if err != nil {
		return false, fmt.Errorf("error fetching API groups: %v", err)
	}

	for _, group := range apiGroups.Groups {
		if group.Name == utils.RouteAPIGroupName {
			return true, nil
		}
	}

	return false, nil

}
