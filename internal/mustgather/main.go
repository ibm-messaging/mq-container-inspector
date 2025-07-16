package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/namespace"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/tarzip"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func MustGather(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the required clients
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	dynamicClient, err := kubeclient.BuildKubernetesDynamicClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building dynamic client from config: %v", err)
	}

	routeClient, err := kubeclient.BuildRouteClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building route client: %v", err)
	}

	// check if the provided queue manager namespace exists on the cluster
	namespaceExists, err := namespace.DoesNamespacesExist(coreClient, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while checking if the namespace %v exists: %v", flags.QueueManagerNamespace, err)
	} else if !namespaceExists {
		return fmt.Errorf("provided queue manager namespace %v was not found on the currently logged-in cluster", flags.QueueManagerNamespace)
	}

	// collect cr must-gathers
	err = gatherCRsToFiles(dynamicClient, flags)
	if err != nil {
		return err
	}

	// collect pods must-gathers
	err = gatherPodsToFiles(coreClient, flags)
	if err != nil {
		return err
	}

	// collect the route must-gathers
	err = gatherRoutesToFiles(cfg, routeClient, flags)
	if err != nil {
		return err
	}

	// collect StatefulSet must-gathers
	err = gatherStatefulSetToFiles(coreClient, flags)
	if err != nil {
		return err
	}

	// collect Service must-gathers
	err = gatherServicesToFiles(coreClient, flags)
	if err != nil {
		return err
	}

	// collect PVC must-gathers
	err = gatherPVCToFiles(coreClient, flags)
	if err != nil {
		return err
	}

	// collect MQ operator must-gathers
	err = gatherMQOperatorToFiles(coreClient, dynamicClient, flags)
	if err != nil {
		return err
	}

	// collect the cp4i csv details
	err = gatherCp4IOperatorCSVToFiles(dynamicClient, flags)
	if err != nil {
		return err
	}

	// collect web-console logs
	err = gatherMQWebConsoleLogsToFiles(cfg, coreClient, flags)
	if err != nil {
		return err
	}

	// collect runmqras logs
	err = gatherRunmqrasLogToFiles(cfg, coreClient, flags)
	if err != nil {
		return err
	}

	// if tar is enabled, then zip the must-gather output
	if flags.TarZip {
		if err := tarzip.TarZipFolder(flags.OutputDir); err != nil {
			return fmt.Errorf("error tar zipping the collected must-gather at path %s: %v", flags.OutputDir, err)
		}
	}

	return nil

}
