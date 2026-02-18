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

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"
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

	var podList []corev1.Pod
	var err error

	if flags.QueueManagerName != "" {
		qmLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

		// Get the queue manager pods
		podList, err = pods.GetPodsBySelector(coreClient, qmLabelSelector, flags.QueueManagerNamespace)
		if err != nil {
			return err
		}
	} else if flags.PodName != "" {

		// Get the queueManager pods from podName
		podList, err = pods.GetMQReplicaPodsViaService(coreClient, flags.PodName, flags.QueueManagerNamespace)
		if err != nil {
			return err
		}
	}

	for _, pod := range podList {

		// Copy console.log from queue manager container to OutputDir
		copyConfig := container.CopyConfig{
			SourcePath:    utils.WebConsoleLogPath,
			KubeConfig:    cfg,
			PodName:       pod.Name,
			Namespace:     pod.Namespace,
			ContainerName: flags.QmContainerName,
			Logger:        logger,
		}

		consoleOutputFilePath := filepath.Join(webconsoleDirectory, fmt.Sprintf("web-%s-console.log", copyConfig.PodName))

		if err := container.CopyPathToFile(copyConfig, consoleOutputFilePath, 10); err != nil {
			logger.Error(fmt.Sprintf("unable to copy web-%s-console.log for pod %q: %v\n", copyConfig.PodName, pod.ObjectMeta.Name, err))
		}

		// Copy messages.log from queue manager container to OutputDir
		copyConfig = container.CopyConfig{
			SourcePath:    utils.WebConsoleMessageLogPath,
			KubeConfig:    cfg,
			PodName:       pod.Name,
			Namespace:     pod.Namespace,
			ContainerName: flags.QmContainerName,
			Logger:        logger,
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
