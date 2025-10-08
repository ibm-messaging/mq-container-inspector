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
package runmqras

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecRunmqrasBySelector exec the runmqras command for every pod in a given namespace that match the provided label selector.
// Parameters:
//   - cfg: the kubernetes config
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the pods.
//   - namespace: the namespace in which to search for the pods.
//   - baseContainer: the container inside which to run the runmqras.
//   - useCustomFile: tells the runmqras tool that it will be run using a custom file
func ExecRunmqrasBySelector(cfg *rest.Config, client kubernetes.Interface, podList []corev1.Pod, namespace, baseContainer string, logger *slog.Logger, useCustomFile bool) ([]container.CopyConfig, error) {

	var runmqrasPodExecutorList []container.CopyConfig

	for _, pod := range podList {
		copyConfig, err := execRunmqras(cfg, pod, useCustomFile, baseContainer)
		if err != nil {
			// log the error, and continue trying to run the runmqras for other pods in podList
			logger.Error(fmt.Sprintf("error while executing runmqras inside pod %s: %v", pod.Name, err))
			continue
		}
		runmqrasPodExecutorList = append(runmqrasPodExecutorList, copyConfig)
	}

	return runmqrasPodExecutorList, nil
}

func execRunmqras(cfg *rest.Config, pod corev1.Pod, useCustomFile bool, baseContainer string) (container.CopyConfig, error) {

	currentTimestamp := utils.GetCurrentTimestamp(utils.TimestampFormat)
	workDir := fmt.Sprintf("/tmp/runmqras_%s", currentTimestamp)

	cmd := []string{
		"runmqras",
		"-workdirectory", workDir,
	}

	if useCustomFile {
		cmd = append(cmd, "-inputfile", fmt.Sprintf("/run/%s", utils.CustomISAFileName))
	} else {
		cmd = append(cmd, "-section", "logger,mqweb,nativeha,trace")
	}

	runmqras := utils.ExecConfig{
		KubernetesConfig: cfg,
		PodName:          pod.Name,
		Namespace:        pod.Namespace,
		ContainerName:    baseContainer,
		Cmd:              cmd,
	}
	execRequestExecutor, err := pods.ExecCmd(runmqras)
	if err != nil {
		return container.CopyConfig{}, fmt.Errorf("error while executing cmd(%#v) inside pod %s: %v", cmd, pod.Name, err)
	}

	err = execRequestExecutor.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: io.Discard,
	})
	if err != nil {
		return container.CopyConfig{}, err
	}

	// create CopyConfig to be used for copying the runmqras details from the container
	writeContainerExecConfig := container.NewContainerCopyConfig(workDir, runmqras)
	return writeContainerExecConfig, nil
}
