package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pvc"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherPVCToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the kubernetes client from config
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	pvcLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get pvc's by selector
	pvcList, err := pvc.GetPVCDetailsBySelector(coreClient, pvcLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching pvc's with selector %s: %v", pvcLabelSelector, err)
	}

	// write the pvc's to their respective yaml files
	pvcYamlFileNameFormat := "%s-pvc.yaml"
	if err := pvc.WritePVCYamlsToFile(pvcList, pvcYamlFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	return nil

}
