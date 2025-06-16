package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/crd"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherCrdToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the kubernetes dynamic client from config
	dynamicClient, err := kubeclient.BuildKubernetesDynamicClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building dynamic client from config: %v", err)
	}

	// get QueueManager details by QueueManager name
	queueManagerDetailsMap, err := crd.GetQueueManagerCrdDetailsByName(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error retrieving queue-manager %s in the namespace %s: %v", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	//write the QueueManager details to its yaml
	fileNameFormat := "%s.yaml"
	if err := crd.WriteQueueManagerCrdYamlToFiles(queueManagerDetailsMap, fileNameFormat, flags.OutputDir, flags.QueueManagerName); err != nil {
		return err
	}

	return nil

}
