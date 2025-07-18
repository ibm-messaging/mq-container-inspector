package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Copies the MQ Webconsole console.log and messages.log to the must gather OutputDir
func gatherMQWebConsoleLogsToFiles(cfg *rest.Config, coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create webconsole directory to store webconsole files
	webconsoleDirectory := filepath.Join(flags.OutputDir, "webconsole")
	directoryExists := utils.CheckIfDirectoryExist(webconsoleDirectory)
	if !directoryExists {
		if err := utils.CreateDirectory(webconsoleDirectory, 0775); err != nil {
			return err
		}
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

		consoleOutputFilePath := filepath.Join(webconsoleDirectory, fmt.Sprintf("web-%s-console.log", copyConfig.PodName))

		if err := container.CopyPathToFile(copyConfig, consoleOutputFilePath, 10); err != nil {
			logger.Error(fmt.Sprintf("unable to copy console.log for pod %q: %v\n", pod.Name, err))
		}

		// Copy messages.log from queue manager container to OutputDir
		copyConfig = container.CopyConfig{
			SourcePath:    utils.WebConsoleMessageLogPath,
			KubeConfig:    cfg,
			PodName:       pod.Name,
			Namespace:     pod.Namespace,
			ContainerName: utils.QmgrContainer,
		}

		messagesOutputFilePath := filepath.Join(webconsoleDirectory, fmt.Sprintf("web-%s-messages.log", copyConfig.PodName))

		if err := container.CopyPathToFile(copyConfig, messagesOutputFilePath, 10); err != nil {
			logger.Error(fmt.Sprintf("unable to copy messages.log for pod %q: %v\n", pod.Name, err))
		}
	}

	if fileCount, err := utils.GetFileCountInDirectory(webconsoleDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("WebConsole Details: %s: Total Files: %d", webconsoleDirectory, fileCount))
	}

	return nil

}
