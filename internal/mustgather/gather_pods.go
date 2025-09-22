/*
© Copyright IBM Corporation 2025

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
package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func gatherPodsToFiles(client kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) ([]string, error) {

	// create pods directory to store pod files
	podsDirectory := filepath.Join(flags.OutputDir, "pods")
	directoryExist := utils.CheckIfDirectoryExist(podsDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(podsDirectory, 0775); err != nil {
			return nil, err
		}
	}

	var podList []corev1.Pod
	var podLogs map[string]utils.PodLogs
	var podDescribeLogs map[string]string
	var podEvents map[string][]corev1.Event
	var err error

	if flags.QueueManagerName != "" {

		podList, podLogs, podDescribeLogs, podEvents, err = getPodDetailsBySelector(client, flags, logger)
		if err != nil {
			return nil, err
		}

	} else if flags.PodName != "" {

		podList, podLogs, podDescribeLogs, podEvents, err = getPodDetailsByName(client, flags, logger)
		if err != nil {
			return nil, err
		}

	}

	// check and collect if we have any failed pods
	var failedPodNames []string
	for _, pod := range podList {
		if pods.GetPodStatus(pod) != utils.PodRunningStatus {
			failedPodNames = append(failedPodNames, pod.ObjectMeta.Name)
		}
	}

	// write the pod details to a file
	podDetailsFileNameFormat := "pod-details.txt"
	if err := pods.WritePodDetailsToFile(podList, podDetailsFileNameFormat, podsDirectory); err != nil {
		return nil, err
	}

	// write the pod details to their respective yamls
	podYamlFileNameFormat := "%s.yaml"
	if err := pods.WritePodYamlsToFile(podList, podYamlFileNameFormat, podsDirectory); err != nil {
		return nil, err
	}

	// write the pods logs to their files
	podLogsFileNameFormat := "%s-%s-pod-log.txt"
	if err := pods.WritePodLogsToFile(podLogs, podLogsFileNameFormat, podsDirectory); err != nil {
		return nil, err
	}

	// write the pod describe logs to their files
	podDescribeLogsFileNameFormat := "%s-describe-log.txt"
	if err := pods.WritePodDescribeLogsToFile(podDescribeLogs, podDescribeLogsFileNameFormat, podsDirectory); err != nil {
		return nil, err
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

	return failedPodNames, nil
}

// getPodDetailsBySelector retrieves pod details (List, Logs, Describe, and Events)
// for pods in the Kubernetes cluster matching the provided selector.
// If no resources are found for a given category, the corresponding return value will be nil.
// Returns an error if the retrieval process fails.
func getPodDetailsBySelector(client kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) ([]corev1.Pod, map[string]utils.PodLogs, map[string]string, map[string][]corev1.Event, error) {
	podLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get pods by selector
	pods, err := pods.GetPodsBySelector(client, podLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching pods with selector %s: %v", podLabelSelector, err))
		return nil, nil, nil, nil, fmt.Errorf("error while fetching pods with selector %s: %v", podLabelSelector, err)
	}

	return getPodDetailsFromPods(client, pods, flags, logger)
}

// getPodDetailsByName retrieves pod details (List, Logs, Describe, and Events)
// for the specified pod name in the Kubernetes cluster. It also discovers any replicas of the pod,
// which may exist in multi-instance or native HA configurations, and gathers their details.
// If no resources are found for a given category, the corresponding return value will be nil.
// Returns an error if the retrieval process fails.
func getPodDetailsByName(client kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) ([]corev1.Pod, map[string]utils.PodLogs, map[string]string, map[string][]corev1.Event, error) {

	// fetch the podNames from the matching service
	pods, err := pods.GetMQReplicaPodsViaService(client, flags.PodName, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching pods with name %s: %v", flags.PodName, err))
		return nil, nil, nil, nil, fmt.Errorf("error while fetching pods with name %s: %v", flags.PodName, err)
	}

	return getPodDetailsFromPods(client, pods, flags, logger)
}

func getPodDetailsFromPods(client kubernetes.Interface, podList []corev1.Pod, flags utils.MustGatherFlags, logger *slog.Logger) ([]corev1.Pod, map[string]utils.PodLogs, map[string]string, map[string][]corev1.Event, error) {

	logger.Info(fmt.Sprintf("Found %d pods in %s namespace with %s name", len(podList), flags.QueueManagerNamespace, flags.PodName))

	// fetch the pod logs
	podLogs := pods.GetPodLogs(client, podList, flags.QueueManagerNamespace, utils.QmgrContainer)

	// fetch the pod describe logs
	podDescribeLogs, err := pods.GetPodDescribeLogs(client, podList, flags.QueueManagerNamespace)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error while fetching pod describe logs with name %s: %v", flags.PodName, err)
	}

	// fetch the podEvents
	podEvents, err := pods.GetPodEvents(client, podList, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching pod events with name %s: %v", flags.PodName, err))
	}

	return podList, podLogs, podDescribeLogs, podEvents, nil
}
