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

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/runmqras"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func gatherRunmqrasLogToFiles(cfg *rest.Config, coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	var podList []corev1.Pod
	var err error

	if flags.QueueManagerName != "" {
		runmqrasLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

		// get the queuemanager pods by label selector
		podList, err = pods.GetPodsBySelector(coreClient, runmqrasLabelSelector, flags.QueueManagerNamespace)
		if err != nil {
			logger.Error(fmt.Sprintf("unable to execute runmqras command in container, the reason being: %v\n", err))
		}
	} else if flags.PodName != "" {

		// get the pods by pod name
		podList, err = pods.GetMQReplicaPodsViaService(coreClient, flags.PodName, flags.QueueManagerNamespace)
		if err != nil {
			logger.Error(fmt.Sprintf("unable to execute runmqras command in container, the reason being: %v\n", err))
		}

	}

	runmqrasCopyConfigs, err := runmqras.ExecRunmqrasBySelector(cfg, coreClient, podList, flags.QueueManagerNamespace, logger)
	if err != nil {
		// continue in case of error
		logger.Error(fmt.Sprintf("unable to execute runmqras command in container, the reason being: %v\n", err))
	}

	for _, runmqrasCopyConfig := range runmqrasCopyConfigs {
		if err := container.CopyPathToFile(runmqrasCopyConfig, flags.OutputDir, 10); err != nil {
			logger.Error(err.Error())
		}
	}

	return nil

}
