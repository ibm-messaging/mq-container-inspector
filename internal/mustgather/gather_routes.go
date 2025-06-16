package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/routes"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherRoutesToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// since routes are OCP specific, check if routes exist on the cluster
	isRoutePresent, err := checkIfRoutesArePresentInCluster(cfg)
	if err != nil {
		return err
	}

	if !isRoutePresent {
		fmt.Printf("routes %s are not present on this cluster\n", utils.RouteAPIGroupName)
		return nil
	}

	// build a route client
	routeClient, err := kubeclient.BuildRouteClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building route client: %v", err)
	}

	routeLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)
	routeFieldSelector := fmt.Sprintf("spec.to.name=%s-ibm-mq", flags.QueueManagerName)

	// get routes created by mq-operator
	routeListBySelector, err := routes.GetRouteDetailsBySelector(routeClient, routeLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching routes with selector %s: %v", routeLabelSelector, err)
	}

	// write the route details to their respective yamls
	routesDetailsBySelectorFileNameFormat := "%s-routes.yaml"
	if err := routes.WriteRouteDetailsBySelectorToFile(routeListBySelector, routesDetailsBySelectorFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// get routes to the QueueManager
	routeListByFieldSelector, err := routes.GetRouteDetailsByFieldSelector(routeClient, routeFieldSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching routes with field selector %s: %v", routeFieldSelector, err)
	}

	// write the route details to their respective yamls
	routesDetailsByFieldSelectorFileNameFormat := "%s-routes-to-qm.yaml"
	if err := routes.WriteRouteDetailsBySelectorToFile(routeListByFieldSelector, routesDetailsByFieldSelectorFileNameFormat, flags.OutputDir); err != nil {
		return err
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
