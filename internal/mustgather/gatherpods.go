package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/namespace"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherPodsToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the kubernetes client from config
	client, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	// check if QueueManager and MQ Operator namespace exists
	if namespaceName, err := namespace.CheckIfNamespacesExist(client, flags.QueueManagerNamespace, flags.OperatorNamespace); err != nil {
		return fmt.Errorf("error while checking if the namespaces %v exist: %v", namespaceName, err)
	}

	podLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get pods by selector
	podList, err := pods.GetPodsBySelector(client, podLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching pods with selector %s: %v", podLabelSelector, err)
	}

	// write the pod details to a file
	podDetailsFileNameFormat := "pod-details.txt"
	if err := pods.WritePodDetailsToFile(podList, podDetailsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// write the pod details to their respective yamls
	podYamlFileNameFormat := "%s.yaml"
	if err := pods.WritePodYamlsToFile(podList, podYamlFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// get pod logs by selector
	podLogs, err := pods.GetPodLogsBySelector(client, podLabelSelector, flags.QueueManagerNamespace, "qmgr")
	if err != nil {
		return fmt.Errorf("error while fetching pod logs with selector %s: %v", podLabelSelector, err)
	}

	// write the pods logs to their files
	podLogsFileNameFormat := "%s-%s-pod-log.txt"
	if err := pods.WritePodLogsToFile(podLogs, podLogsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// get pod describe logs by selector
	podDescribeLogs, err := pods.GetPodDescribeBySelector(client, podLabelSelector, flags.QueueManagerNamespace, true)
	if err != nil {
		return fmt.Errorf("error while fetching pod describe logs with selector %s: %v", podLabelSelector, err)
	}

	// write the pod describe logs to their files
	podDescribeLogsFileNameFormat := "%s-describe-log.txt"
	if err := pods.WritePodDescribeLogsToFile(podDescribeLogs, podDescribeLogsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// get pod events by selector
	podEvents, err := pods.GetPodEventsBySelector(client, podLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching pod events with selector %s: %v", podLabelSelector, err)
	}

	// write pod events to their files
	podEventsFileNameFormat := "%s-pod-events.txt"
	if err := pods.WritePodEventsToFile(podEvents, podEventsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	return nil
}
