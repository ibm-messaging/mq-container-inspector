package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/runmqras"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherRunmqrasLogToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the kubernetes core client from config
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	runmqrasLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	runmqrasCopyConfigs, err := runmqras.ExecRunmqrasBySelector(cfg, coreClient, runmqrasLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while executing runmqras command in container: %v", err)
	}

	for _, runmqrasCopyConfig := range runmqrasCopyConfigs {
		if err := container.CopyPathToFile(runmqrasCopyConfig, flags.OutputDir, 10); err != nil {
			return err
		}
	}

	return nil

}
