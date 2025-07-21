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
