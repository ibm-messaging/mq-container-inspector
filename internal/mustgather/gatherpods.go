package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherPodsToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the kubernetes client from config
	client := kubeclient.BuildKubernetesClientFromConfig(cfg)

	podLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get pods by selector
	podList, err := pods.GetPodsBySelector(client, podLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("Error while fetching pods with selector %s: %v", podLabelSelector, err)
	}
	fmt.Printf("(%d) Pods retrieved: %#v\n", len(podList), podList)

	// TODO: Use the get pods output to write to a YAML file

	// get pod logs by selector
	podLogs, err := pods.GetPodLogsBySelector(client, podLabelSelector, flags.QueueManagerNamespace, "qmgr")
	if err != nil {
		return fmt.Errorf("Error while fetching pod logs with selector %s: %v", podLabelSelector, err)
	}
	fmt.Printf("Pods logs retrieved: %v\n", podLogs)

	// get pod describe logs by selector
	podDescribeLogs, err := pods.GetPodDescribeBySelector(client, podLabelSelector, flags.QueueManagerNamespace, true)
	if err != nil {
		return fmt.Errorf("Error while fetching pod describe logs with selector %s: %v", podLabelSelector, err)
	}
	fmt.Printf("Pods Describe logs retrieved: %v\n", podDescribeLogs)

	// get pod events by selector
	podEvents, err := pods.GetPodEventsBySelector(client, podLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("Error while fetching pod events with selector %s: %v", podLabelSelector, err)
	}
	fmt.Printf("Pods events retrieved: %v\n", podEvents)

	return nil
}
