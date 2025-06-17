package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/service"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherServicesToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the kubernetes core client
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	serviceLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get the service details by selector
	serviceList, err := service.GetServiceDetailsBySelector(coreClient, serviceLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching services with selector %s: %v", serviceLabelSelector, err)
	}

	// write the services to their respective yaml files
	serviceDetailsFileNameFormat := "%s-service.yaml"
	if err := service.WriteServiceYamlsToFile(serviceList, serviceDetailsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	return nil

}
