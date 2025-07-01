package pods

import (
	"fmt"
	"net/http"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func ExecCmd(config utils.ExecConfig) (remotecommand.Executor, error) {

	// build the kubernetes core client from config
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(config.KubernetesConfig)
	if err != nil {
		return nil, fmt.Errorf("error building core client from config: %v", err)
	}

	execRequest := coreClient.CoreV1().RESTClient().Post().Resource(utils.ResourcePod).Name(config.PodName).Namespace(config.Namespace).SubResource(utils.ExecSubcommand).VersionedParams(&corev1.PodExecOptions{
		Container: config.ContainerName,
		Command:   config.Cmd,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	execRequestExecutor, err := remotecommand.NewSPDYExecutor(config.KubernetesConfig, http.MethodPost, execRequest.URL())
	if err != nil {
		return nil, err
	}

	return execRequestExecutor, nil

}
