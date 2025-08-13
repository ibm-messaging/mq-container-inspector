package mustgather

import (
	"context"
	"fmt"
	"os"
	"slices"
	"text/tabwriter"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/cr"
	ns "github.ibm.com/mq-cloudpak/mq-inspector/pkg/namespace"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func validateNamespace(coreClient kubernetes.Interface, namespace string) error {
	namespaceExists, err := ns.DoesNamespacesExist(coreClient, namespace)
	if err != nil {
		return fmt.Errorf("error while checking if the namespace %v exists: %v", namespace, err)
	} else if !namespaceExists {
		return fmt.Errorf("provided queue manager namespace %v was not found on the currently logged-in cluster", namespace)
	}

	return nil
}

func validatePodName(coreClient kubernetes.Interface, flags *utils.MustGatherFlags) (*corev1.Pod, error) {

	if flags.PodName == "" {
		return nil, nil
	}

	pod, err := pods.GetPodByName(coreClient, flags.PodName, flags.QueueManagerNamespace)
	if errors.IsNotFound(err) {
		fmt.Printf("Provided pod %s, is not present in the %s namespace. Please select a pod from the following list:\n", flags.PodName, flags.QueueManagerNamespace)
		if podNames, err := listAllQueueManagerPods(coreClient, flags.QueueManagerNamespace); err != nil {
			return nil, err
		} else if podNames == nil {
			fmt.Printf("No queue manager pods found in the %s namespace\n", flags.QueueManagerNamespace)
		} else {
			printQueueManagerPodNames(podNames)
		}
		return nil, fmt.Errorf("re-run the must-gather tool with the correct queue manager pod name")
	} else if err != nil {
		return nil, fmt.Errorf("error while checking instance label for pod %s: %v", flags.PodName, err)
	}

	// check if the pod is a queue manager pod
	if !isQueueManagerPod(pod) {
		// display all the queue manager pods in the namespace
		fmt.Printf("Provided pod %s, is not a queue manager pod. Please select a pod from the following list:\n", flags.PodName)
		podNames, err := listAllQueueManagerPods(coreClient, flags.QueueManagerNamespace)
		if err != nil {
			return pod, err
		} else if podNames == nil {
			return pod, fmt.Errorf("no queue manager pods found in the %s namespace", flags.QueueManagerNamespace)
		} else {
			printQueueManagerPodNames(podNames)
		}
		return pod, fmt.Errorf("re-run the must-gather tool with the correct queue manager pod name")
	}

	return pod, nil

}

func validateQueueManagerName(coreClient kubernetes.Interface, dynamicClient dynamic.Interface, flags *utils.MustGatherFlags) (*corev1.Pod, error) {

	if flags.QueueManagerName == "" {
		return nil, nil
	}

	// check if we have pods with the instance as --qm-name
	labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)
	podList, err := pods.GetPodsBySelector(coreClient, labelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return nil, fmt.Errorf("error while fetching pods with selector %s in namespace %s", labelSelector, flags.QueueManagerNamespace)
	} else if podList == nil {
		// display all the QMGR pods in the namespace
		fmt.Printf("Queue manager '%s' not found in the namespace '%s'\n", flags.QueueManagerName, flags.QueueManagerNamespace)
		if err := listAllQueueManagerCRs(dynamicClient, flags); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("no queue manager found with name %s in namespace %s", flags.QueueManagerName, flags.QueueManagerNamespace)
	}

	return &podList[0], nil

}

func listAllQueueManagerCRs(dynamicClient dynamic.Interface, flags *utils.MustGatherFlags) error {

	queueManagerList, err := cr.ListQueueManagersInNamespace(dynamicClient, flags.QueueManagerNamespace)
	if errors.IsNotFound(err) || len(queueManagerList) == 0 {
		fmt.Printf("No queue managers custom resorces found in the namespace '%s'. Please validate your namespace is correct\n", flags.QueueManagerNamespace)
		return fmt.Errorf("queue manager '%s' not found", flags.QueueManagerName)
	} else if err != nil {
		return fmt.Errorf("error retrieving queue managers custom resorces in the namespace %s: %v", flags.QueueManagerNamespace, err)
	}

	fmt.Printf("Available queue managers in namespace '%s':\n", flags.QueueManagerNamespace)
	printQueueManagerCRs(queueManagerList)

	return nil

}

func isQueueManagerPod(pod *corev1.Pod) bool {

	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == utils.QueueManagerEnvName && env.Value != "" {
				return true
			}
		}
	}

	return false

}

func listAllQueueManagerPods(coreClient kubernetes.Interface, namespace string) ([]string, error) {

	podList, err := coreClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var queueManagerPodNames []string

	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			for _, env := range container.Env {
				if env.Name == utils.QueueManagerEnvName && env.Value != "" {
					queueManagerPodNames = append(queueManagerPodNames, pod.Name)
				}
			}
		}
	}

	return slices.Compact(queueManagerPodNames), nil

}

func printQueueManagerPodNames(queueManagerPodNames []string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(writer, "NAME")

	for _, podName := range queueManagerPodNames {
		fmt.Fprintf(writer, "%s\n", podName)
	}

	writer.Flush()

}

func printQueueManagerCRs(queueManagerList []unstructured.Unstructured) {

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// print header
	fmt.Fprintln(writer, "NAME\tPHASE")

	for _, queueManager := range queueManagerList {
		name := queueManager.GetName()
		phase, found, err := unstructured.NestedString(queueManager.Object, "status", "phase")
		if err != nil || !found {
			phase = "<unknown>"
		}
		fmt.Fprintf(writer, "%s\t%s\n", name, phase)
	}

	writer.Flush()

}
