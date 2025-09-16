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
package utils

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type MustGatherFlags struct {
	QueueManagerName      string // QueueManager resource name
	PodName               string // pod name
	QueueManagerNamespace string // QueueManager  resource namespace
	OperatorNamespace     string // namespace where the MQ Operator is running
	KubeconfigPath        string // path to the kubeconfig file
	OutputDir             string // directory where the must-gathers will be stored
	NoTar                 bool   // whether to tar+zip the must-gather output
	NoExec                bool   //whether to collect the pod-exec must-gathers like runmqras etc.
}

type PVCInspectorFlags struct {
	QueueManagerName      string // QueueManager resource name
	PodName               string // pod name
	QueueManagerNamespace string // QueueManager  resource namespace
	KubeconfigPath        string // path to the kubeconfig file
	OutputDir             string // directory where the collected pvc-inspector data will be stored
	Cleanup               bool   // whether to cleanup the pvc-pods at the end of the tool run
	DryRun                bool   // whether to create the pvc-pods
	NoTar                 bool   // whether to tar+zip the pvc-inspector output
}

func GetCurrentTimestamp(timeFormat string) string {
	return time.Now().Format(timeFormat)
}

type PodLogs struct {
	PreviousPodLogsRequest *rest.Request
	CurrentPodLogsRequest  *rest.Request
}

func FetchQMGRResourceNameFromSelector(selector string) (string, error) {
	if index := strings.Index(selector, "="); index != -1 {
		return selector[index+1:], nil
	}

	return "", fmt.Errorf("error invalid selector format: %s", selector)
}

func FormatFilePath(outputDir, fileNameFormat string, args ...interface{}) string {
	filename := fmt.Sprintf(fileNameFormat, args...)
	return filepath.Join(outputDir, filename)
}

type ExecConfig struct {
	KubernetesConfig *rest.Config
	PodName          string
	Namespace        string
	ContainerName    string
	Cmd              []string
}

func CheckIfDirectoryExist(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func CreateDirectory(dir string, perm fs.FileMode) error {
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("error creating output-directory(%s): %v", dir, err)
	}
	return nil
}

func GetLogFilePath(outputDirectory, logFileName string) string {
	return filepath.Join(outputDirectory, logFileName)
}

func InitializeLogFile(outputDirectory, logFileName string) (*os.File, error) {
	logFilePath := GetLogFilePath(outputDirectory, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return nil, fmt.Errorf("error creating logfile %s in base-directory %s: %v", logFileName, outputDirectory, err)
	}
	return logFile, nil

}

func GetFileCountInDirectory(directoryName string) (int, error) {

	entries, err := os.ReadDir(directoryName)
	if err != nil {
		return 0, fmt.Errorf("error while reading %s directory: %v", directoryName, err)
	}

	fileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			fileCount++
		}
	}

	return fileCount, nil

}

func DeleteEmptyDirectories(directoryName string, logger *slog.Logger) error {

	entries, err := os.ReadDir(directoryName)
	if err != nil {
		return fmt.Errorf("error while reading %s directory", directoryName)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			entryPath := filepath.Join(directoryName, entry.Name())
			fileCount, err := GetFileCountInDirectory(entryPath)
			if err != nil {
				return err
			}

			if fileCount == 0 {
				if err := os.RemoveAll(entryPath); err != nil {
					return err
				} else {
					logger.Info(fmt.Sprintf("Deleting empty directory: %s/%s", directoryName, entry.Name()))
				}
			}
		}
	}

	return nil

}

func GetPodInstance(pod *corev1.Pod) string {

	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == NativeHAEnvName && env.Value == "true" {
				return NativeHA
			} else if env.Name == MultiInstanceEnvName && env.Value == "true" {
				return MultiInstance
			}
		}
	}
	return SingleInstance
}

func IsSelectorSubsetOfLabels(selector, label map[string]string) bool {

	for key, value := range selector {
		if label[key] != value {
			return false
		}
	}

	return true

}

func GetPodOwner(pod *corev1.Pod) string {

	// a pod can have multiple owners but only a single controlling owner which is what we return
	for _, owner := range pod.ObjectMeta.OwnerReferences {
		if owner.Kind != "" && *owner.Controller {
			return owner.Kind
		}
	}

	return ""
}

func GetPVCInspectorContainerCommand() []string {
	return []string{"tail", "-f", "/dev/null"}
}

func CheckIfPodHasPersistedStorage(pod corev1.Pod) bool {

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			return true
		}
	}

	return false

}
