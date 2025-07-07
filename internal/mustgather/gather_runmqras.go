package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/runmqras"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func gatherRunmqrasLogToFiles(cfg *rest.Config, coreClient kubernetes.Interface, flags utils.MustGatherFlags) error {

	runmqrasLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	runmqrasCopyConfigs, err := runmqras.ExecRunmqrasBySelector(cfg, coreClient, runmqrasLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		// continue in case of error
		fmt.Printf("unable to execute runmqras command in container, the reason being:: %v\n", err)
	}

	for _, runmqrasCopyConfig := range runmqrasCopyConfigs {
		if err := container.CopyPathToFile(runmqrasCopyConfig, flags.OutputDir, 10); err != nil {
			return err
		}
	}

	return nil

}
