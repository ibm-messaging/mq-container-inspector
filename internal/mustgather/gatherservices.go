package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/service"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
)

func gatherServicesToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags) error {

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
