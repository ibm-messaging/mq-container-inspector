/*
© Copyright IBM Corporation 2026

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
	"fmt"
	"strings"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/cr"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func CheckIfPDBEnabled(dynamicClient dynamic.Interface, pod *corev1.Pod, qmName, qmNamespace, containerName string) (bool, error) {

	if pod != nil {

		if containerEnvEquals(pod, containerName, utils.MQPDBEnvVariable, "true") {
			return true, nil
		}

		if podHasMetadataServerSidecar(pod) {
			return true, nil
		}

		if _, ok := pod.Labels[utils.MQPDBLabel]; ok {
			return true, nil
		}
	}

	// fallback to qm CR
	if qmName == "" || qmNamespace == "" {
		return false, fmt.Errorf("queueManager name and queueManager namespace should not be empty")
	}

	qm, err := cr.GetQueueManagerCrDetailsByName(dynamicClient, qmName, qmNamespace)
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	enabled, err := checkQMCRIfPDBEnabled(qm)
	if err != nil {
		return false, fmt.Errorf("queueManager %s/%s: %w", qmNamespace, qmName, err)
	}
	return enabled, nil

}

func checkQMCRIfPDBEnabled(qm map[string]interface{}) (bool, error) {

	enabled, found, err := unstructured.NestedBool(
		qm,
		"spec", "queueManager", "availability", "podDisruptionBudget", "enabled",
	)
	if err != nil {
		return false, fmt.Errorf("parsing spec.queueManager.availability.podDisruptionBudget.enabled: %w", err)
	}

	if !found {
		return false, nil
	}

	return enabled, nil

}

func podHasMetadataServerSidecar(pod *corev1.Pod) bool {

	for _, container := range pod.Spec.InitContainers {

		if container.RestartPolicy == nil || *container.RestartPolicy != corev1.ContainerRestartPolicyAlways {
			continue
		}

		if container.Name == utils.MetadataServerContainerName {
			return true
		}

		// the container may have been renamed, fall back to the command it runs
		for _, arg := range container.Command {
			if strings.Contains(arg, utils.MetadataServerCommand) {
				return true
			}
		}
	}

	return false

}

func containerEnvEquals(pod *corev1.Pod, containerName, envName, value string) bool {

	for _, container := range pod.Spec.Containers {

		if containerName != "" && container.Name != containerName {
			continue
		}

		for _, env := range container.Env {
			if env.Name == envName && env.Value == value {
				return true
			}
		}
	}

	return false

}
