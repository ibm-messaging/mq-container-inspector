package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
)

func gatherPodsToFiles(client kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create pods directory to store pod files
	podsDirectory := filepath.Join(flags.OutputDir, "pods")
	directoryExist := utils.CheckIfDirectoryExist(podsDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(podsDirectory, 0775); err != nil {
			return err
		}
	}

	podLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get pods by selector
	podList, err := pods.GetPodsBySelector(client, podLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching pods with selector %s: %v", podLabelSelector, err))
		return fmt.Errorf("error while fetching pods with selector %s: %v", podLabelSelector, err)
	}

	logger.Info(fmt.Sprintf("Found %d pods in %s namespace with %s label", len(podList), flags.QueueManagerNamespace, podLabelSelector))

	// write the pod details to a file
	podDetailsFileNameFormat := "pod-details.txt"
	if err := pods.WritePodDetailsToFile(podList, podDetailsFileNameFormat, podsDirectory); err != nil {
		return err
	}

	// write the pod details to their respective yamls
	podYamlFileNameFormat := "%s.yaml"
	if err := pods.WritePodYamlsToFile(podList, podYamlFileNameFormat, podsDirectory); err != nil {
		return err
	}

	// get pod logs by selector
	podLogs, err := pods.GetPodLogsBySelector(client, podLabelSelector, flags.QueueManagerNamespace, utils.QmgrContainer)
	if err != nil {
		return fmt.Errorf("error while fetching pod logs with selector %s: %v", podLabelSelector, err)
	}

	// write the pods logs to their files
	podLogsFileNameFormat := "%s-%s-pod-log.txt"
	if err := pods.WritePodLogsToFile(podLogs, podLogsFileNameFormat, podsDirectory); err != nil {
		return err
	}

	// get pod describe logs by selector
	podDescribeLogs, err := pods.GetPodDescribeBySelector(client, podLabelSelector, flags.QueueManagerNamespace, true)
	if err != nil {
		return fmt.Errorf("error while fetching pod describe logs with selector %s: %v", podLabelSelector, err)
	}

	// write the pod describe logs to their files
	podDescribeLogsFileNameFormat := "%s-describe-log.txt"
	if err := pods.WritePodDescribeLogsToFile(podDescribeLogs, podDescribeLogsFileNameFormat, podsDirectory); err != nil {
		return err
	}

	// get pod events by selector
	podEvents, err := pods.GetPodEventsBySelector(client, podLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching pod events with selector %s: %v", podLabelSelector, err))
	}

	// write pod events to their files
	podEventsFileNameFormat := "%s-pod-events.txt"
	if err := pods.WritePodEventsToFile(podEvents, podEventsFileNameFormat, podsDirectory); err != nil {
		logger.Error(err.Error())
	}

	if fileCount, err := utils.GetFileCountInDirectory(podsDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Pod details: %s: Total Files: %d", podsDirectory, fileCount))
	}

	return nil
}
