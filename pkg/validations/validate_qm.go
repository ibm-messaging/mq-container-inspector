/*
© Copyright IBM Corporation 2025, 2026

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validations

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/cr"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	ns "github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/namespace"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateNamespace(coreClient kubernetes.Interface, namespace string) error {
	namespaceExists, err := ns.DoesNamespacesExist(coreClient, namespace)
	if err != nil {
		if errors.IsUnauthorized(err) {
			return fmt.Errorf("unauthorized access to namespace: %v. Please ensure you are logged in and have the necessary access rights for the namespace", namespace)
		}
		return fmt.Errorf("error while checking if the namespace %v exists: %v", namespace, err)
	} else if !namespaceExists {
		return fmt.Errorf("provided namespace %v was not found on the currently logged-in cluster", namespace)
	}

	return nil
}

func ValidatePodName(coreClient kubernetes.Interface, podName, namespace string) (*corev1.Pod, error) {

	if podName == "" {
		return nil, nil
	}

	pod, err := pods.GetPodByName(coreClient, podName, namespace)
	if errors.IsNotFound(err) {
		fmt.Printf("Provided pod %s, is not present in the %s namespace. Please select a pod from the following list:\n", podName, namespace)
		if podNames, err := listAllQueueManagerPods(coreClient, namespace); err != nil {
			return nil, err
		} else if podNames == nil {
			fmt.Printf("No queue manager pods found in the %s namespace\n", namespace)
		} else {
			printQueueManagerPodNames(podNames)
		}
		return nil, fmt.Errorf("re-run the must-gather tool with a valid queue manager pod name")
	} else if err != nil {
		return nil, fmt.Errorf("error while checking instance label for pod %s: %v", podName, err)
	}

	// check if the pod is a queue manager pod
	if !utils.IsQueueManagerPod(pod) {
		// display all the queue manager pods in the namespace
		fmt.Printf("Provided pod %s, is not a queue manager pod. Please select a pod from the following list:\n", podName)
		podNames, err := listAllQueueManagerPods(coreClient, namespace)
		if err != nil {
			return pod, err
		} else if podNames == nil {
			return pod, fmt.Errorf("no queue manager pods found in the %s namespace", namespace)
		} else {
			printQueueManagerPodNames(podNames)
		}
		return pod, fmt.Errorf("re-run the must-gather tool with a valid queue manager pod name")
	}

	return pod, nil

}

func ValidateQueueManagerName(coreClient kubernetes.Interface, dynamicClient dynamic.Interface, qmName, namespace, qmImage string) (*corev1.Pod, error) {

	if qmName == "" {
		return nil, nil
	}

	// check if we have pods with the instance as --qm-name
	labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", qmName)
	podList, err := pods.GetPodsBySelector(coreClient, labelSelector, namespace)
	if err != nil {
		return nil, fmt.Errorf("error while fetching pods with selector %s in namespace %s", labelSelector, namespace)
	} else if podList == nil {
		// check if the queue-manager with qmName exists
		if queueManagerCRExists(dynamicClient, qmName, namespace) {
			return nil, nil
		}

		// check if the fall-back to qm-image is being used
		if qmImage != "" {
			return nil, nil
		}

		// display all the QMGR pods in the namespace
		fmt.Printf("Queue manager '%s' not found in the namespace '%s'\n", qmName, namespace)
		if err := listAllQueueManagerCRs(dynamicClient, qmName, namespace); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("no queue manager found with name %s in namespace %s", qmName, namespace)
	}

	return &podList[0], nil

}

func listAllQueueManagerCRs(dynamicClient dynamic.Interface, qmName, namespace string) error {

	queueManagerList, err := cr.ListQueueManagersInNamespace(dynamicClient, namespace)
	if errors.IsNotFound(err) || len(queueManagerList) == 0 {
		fmt.Printf("No queue managers custom resorces found in the namespace '%s'. Please validate your namespace is correct\n", namespace)
		return fmt.Errorf("queue manager '%s' not found", qmName)
	} else if err != nil {
		return fmt.Errorf("error retrieving queue managers custom resorces in the namespace %s: %v", namespace, err)
	}

	fmt.Printf("Available queue managers in namespace '%s':\n", namespace)
	printQueueManagerCRs(queueManagerList)

	return nil

}

func listAllQueueManagerPods(coreClient kubernetes.Interface, namespace string) (map[string]string, error) {

	podList, err := coreClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	queueManagerPodNames := make(map[string]string)

	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			for _, env := range container.Env {
				if env.Name == utils.QueueManagerEnvName && env.Value != "" {
					queueManagerPodNames[pod.Name] = string(pod.Status.Phase)
				}
			}
		}
	}

	return queueManagerPodNames, nil

}

func printQueueManagerPodNames(podPhaseMap map[string]string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Create a sorted array of podPhaseMap keys
	podNames := make([]string, 0, len(podPhaseMap))
	for k := range podPhaseMap {
		podNames = append(podNames, k)
	}
	sort.Strings(podNames)

	// Print header
	fmt.Fprintln(writer, "NAME\tPHASE")
	// Print the pods and phases in alphabetical order
	for _, podName := range podNames {
		fmt.Fprintf(writer, "%s\t%s\n", podName, podPhaseMap[podName])
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

func queueManagerCRExists(dynamicClient dynamic.Interface, qmName, namespace string) bool {

	if _, err := cr.GetQueueManagerCrDetailsByName(dynamicClient, qmName, namespace); err != nil {
		return false
	}

	return true

}
