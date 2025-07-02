package mustgather

import (
	"fmt"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

// Copies the MQ Webconsole console.log and messages.log to the must gather OutputDir
func gatherMQWebConsoleLogsToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the kubernetes core client from config
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	qmLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// Get the queue manager pods
	podList, err := pods.GetPodsBySelector(coreClient, qmLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return err
	}

	for _, pod := range podList {

		// Copy console.log from queue manager container to OutputDir
		copyConfig := container.CopyConfig{
			SourcePath:    utils.WebConsoleLogPath,
			KubeConfig:    cfg,
			PodName:       pod.Name,
			Namespace:     pod.Namespace,
			ContainerName: utils.QmgrContainer,
		}

		consoleOutputFilePath := filepath.Join(flags.OutputDir, fmt.Sprintf("web-%s-console.log", copyConfig.PodName))

		// TODO: currently this fails and throws an error if the pod is in crashloop backoff.
		if err := container.CopyPathToFile(copyConfig, consoleOutputFilePath, 10); err != nil {
			return err
		}

		// Copy messages.log from queue manager container to OutputDir
		copyConfig = container.CopyConfig{
			SourcePath:    utils.WebConsoleMessageLogPath,
			KubeConfig:    cfg,
			PodName:       pod.Name,
			Namespace:     pod.Namespace,
			ContainerName: utils.QmgrContainer,
		}

		messagesOutputFilePath := filepath.Join(flags.OutputDir, fmt.Sprintf("web-%s-messages.log", copyConfig.PodName))

		if err := container.CopyPathToFile(copyConfig, messagesOutputFilePath, 10); err != nil {
			return err
		}
	}

	return nil

}
