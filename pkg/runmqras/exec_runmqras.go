package runmqras

import (
	"context"
	"fmt"
	"io"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecRunmqrasBySelector exec the runmqras command for every pod in a given namespace that match the provided label selector.
// Parameters:
//   - cfg: the kubernetes config
//   - client: the Kubernetes client used to interact with the cluster.
//   - selector: the label selector used to filter the routes.
//   - namespace: the namespace in which to search for the routes.
func ExecRunmqrasBySelector(cfg *rest.Config, client kubernetes.Interface, selector, namespace string) ([]container.CopyConfig, error) {

	podList, err := pods.GetPodsBySelector(client, selector, namespace)
	if err != nil {
		return nil, err
	}

	var runmqrasPodExecutorList []container.CopyConfig

	for _, pod := range podList {
		copyConfig, err := execRunmqras(cfg, pod)
		if err != nil {
			return runmqrasPodExecutorList, fmt.Errorf("error while executing runmqras inside pod %s: %v", pod.Name, err)
		}
		runmqrasPodExecutorList = append(runmqrasPodExecutorList, copyConfig)
	}

	return runmqrasPodExecutorList, nil
}

func execRunmqras(cfg *rest.Config, pod corev1.Pod) (container.CopyConfig, error) {

	currentTimestamp := utils.GetCurrentTimestamp(utils.TimestampFormat)
	workDir := fmt.Sprintf("/tmp/runmqras_%s", currentTimestamp)

	cmd := []string{
		"runmqras",
		"-workdirectory", workDir,
		"-section", "logger,mqweb,nativeha,trace",
	}
	runmqras := utils.ExecConfig{
		KubernetesConfig: cfg,
		PodName:          pod.Name,
		Namespace:        pod.Namespace,
		ContainerName:    utils.QmgrContainer,
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
