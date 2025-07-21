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
package pods

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// WritePodDetailsToFile writes the details of the given pods to a text file.
// Parameters:
//   - podList:        the list of pods whose details will be written.
//   - fileNameFormat: the format string used to name the output file.
//   - outputDir:      the directory in which the details file will be created.
func WritePodDetailsToFile(podList []corev1.Pod, fileNameFormat, outputDir string) error {

	if len(podList) > 0 {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat)

		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("error creating file(%s): %v", fileName, err)
		}
		defer file.Close()

		// tabwriter will handle dynamic spacing
		writer := tabwriter.NewWriter(file, 0, 8, 2, ' ', 0)

		// write file header
		fmt.Fprintf(writer, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE\tIP\tNODE\n")

		for _, pod := range podList {

			podName := pod.Name
			totalContainers := len(pod.Status.ContainerStatuses)
			readyContainers := getReadyContainersCount(pod)
			ready := fmt.Sprintf("%d/%d", readyContainers, totalContainers)
			podStatus := getPodStatus(pod)
			restartCount := getContainerRestartCount(pod)

			//TODO: where to get the last-restart time like: (17h ago)

			podAge := getPodAge(pod.CreationTimestamp.Time)
			podIP := pod.Status.PodIP
			nodeName := pod.Spec.NodeName

			fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
				podName, ready, podStatus, restartCount, podAge, podIP, nodeName)

		}

		writer.Flush()
	}

	return nil

}

// WritePodYamlsToFile writes each pod’s details to a separate YAML file.
// Parameters:
//   - podList:        the list of pods whose details will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the pod name.
//   - outputDir:      the directory in which the YAML files will be created.
func WritePodYamlsToFile(podList []corev1.Pod, fileNameFormat, outputDir string) error {

	for _, pod := range podList {
		fileName := utils.FormatFilePath(outputDir, fileNameFormat, pod.Name)

		data, err := yaml.Marshal(pod)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for pod %s: %v", pod.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing pod %s data in the yaml file: %v", pod.Name, err)
		}
	}

	return nil

}

// WritePodLogsToFile writes each pod’s logs to their respective files.
// Parameters:
//   - podLogsMap:     the map of pod names to their PodLogs structs, whose logs will be written.
//   - fileNameFormat: the format string used to name each file; must contain two "%s" verbs—first for the pod name, second for "current" or "previous" based on log type.
//   - outputDir:      the directory in which the log files will be created.
func WritePodLogsToFile(podLogsMap map[string]utils.PodLogs, fileNameFormat, outputDir string) error {

	for podName, podLog := range podLogsMap {
		var fileName string

		if podLog.PreviousPodLogsRequest != nil {
			fileName = utils.FormatFilePath(outputDir, fileNameFormat, podName, "previous")

			prevLogFile, err := os.Create(fileName)
			if err != nil {
				return fmt.Errorf("error creating %s pod previous log file %s: %v", podName, fileName, err)
			}

			previousPodLogsStream, err := podLog.PreviousPodLogsRequest.Stream(context.TODO())
			if err != nil {
				return fmt.Errorf("error getting pod %s, previous log stream: %v", podName, err)
			}

			if _, err := io.Copy(prevLogFile, previousPodLogsStream); err != nil {
				prevLogFile.Close()
				return fmt.Errorf("error while writing %s pod previous logs to file %s: %v", podName, fileName, err)
			}

			prevLogFile.Close()

		}

		fileName = utils.FormatFilePath(outputDir, fileNameFormat, podName, "current")

		curLogFile, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("error creating %s pod current log file %s: %v", podName, fileName, err)
		}

		currentPodLogsStream, err := podLog.CurrentPodLogsRequest.Stream(context.TODO())
		if err != nil {
			return fmt.Errorf("error getting pod %s, current log stream: %v", podName, err)
		}

		if _, err := io.Copy(curLogFile, currentPodLogsStream); err != nil {
			curLogFile.Close()
			return fmt.Errorf("error while writing %s pod current logs to file %s: %v", podName, fileName, err)
		}

		curLogFile.Close()
	}

	return nil

}

// WritePodDescribeLogsToFile writes each pod’s describe logs to their respective files.
// Parameters:
//   - podDescribeLogsMap:  the map of pod names to their describe logs, which will be written.
//   - fileNameFormat:      the format string used to name each file; must contain one "%s", which will be replaced by the pod name.
//   - outputDir:           the directory in which the YAML files will be created.
func WritePodDescribeLogsToFile(podDescribeLogsMap map[string]string, fileNameFormat, outputDir string) error {

	for podName, podDescribeLogs := range podDescribeLogsMap {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, podName)

		if err := os.WriteFile(fileName, []byte(podDescribeLogs), 0660); err != nil {
			return fmt.Errorf("error while writing %s pod describe logs to file %s: %v", podName, fileName, err)
		}

	}

	return nil

}

// WritePodEventsToFile writes each pod’s events to their respective files.
// Parameters:
//   - podEventMap:    the map of pod names to their events, which will be written.
//   - fileNameFormat: the format string used to name each file; must contain one "%s", which will be replaced by the pod name.
//   - outputDir:      the directory in which the YAML files will be created.
func WritePodEventsToFile(podEventMap map[string][]corev1.Event, fileNameFormat, outputDir string) error {

	for podName, podEvents := range podEventMap {

		if len(podEvents) > 0 {

			fileName := utils.FormatFilePath(outputDir, fileNameFormat, podName)

			file, err := os.Create(fileName)
			if err != nil {
				return fmt.Errorf("error creating %s pod event file %s: %v", podName, fileName, err)
			}

			// tabwriter will handle dynamic spacing
			writer := tabwriter.NewWriter(file, 0, 8, 2, ' ', 0)

			// write the header
			fmt.Fprintf(writer, "LAST SEEN\tTYPE\tREASON\tKIND\tNAME\tMESSAGE\n")

			for _, event := range podEvents {

				var lastSeen string
				if !event.LastTimestamp.IsZero() {
					lastSeen = event.LastTimestamp.Format(time.RFC3339)
				}

				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
					lastSeen,
					event.Type,
					event.Reason,
					event.InvolvedObject.Kind,
					event.InvolvedObject.Name,
					event.Message)

			}

			writer.Flush()
			file.Close()

		}
	}

	return nil
}

func getReadyContainersCount(pod corev1.Pod) int {
	var readyCount int
	for _, container := range pod.Status.ContainerStatuses {
		if container.Ready {
			readyCount++
		}
	}

	return readyCount
}

func getPodStatus(pod corev1.Pod) string {

	for _, container := range pod.Status.ContainerStatuses {

		if container.State.Waiting != nil {
			return container.State.Waiting.Reason
		} else if container.State.Terminated != nil {
			return container.State.Terminated.Reason
		} else if container.State.Running != nil {
			// we donot have a Reason field for Running status
			return string(pod.Status.Phase)
		}
	}

	return string(pod.Status.Phase)

}

func getContainerRestartCount(pod corev1.Pod) int32 {
	var restartCount int32
	for _, container := range pod.Status.ContainerStatuses {
		restartCount += container.RestartCount
	}
	return restartCount
}

func getPodAge(podCreationTime time.Time) string {
	duration := time.Since(podCreationTime)
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int((duration % time.Hour) / time.Minute)
	seconds := int((duration % time.Minute) / time.Second)
	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}
