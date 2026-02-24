/*
© Copyright IBM Corporation 2025,2026

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
	"slices"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/configmap"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/cr"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

func gatherConfigMapsToFiles(client kubernetes.Interface, dynamicClient dynamic.Interface, qmPod *corev1.Pod, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create configmaps directory to store configmap files
	configMapDirectory := filepath.Join(flags.OutputDir, "configmaps")
	if !utils.CheckIfDirectoryExist(configMapDirectory) {
		if err := utils.CreateDirectory(configMapDirectory, 0o700); err != nil {
			return err
		}
	}

	// fetch the configmap names from the qmPod or queue-manager cr
	var configMapNames []string

	if qmPod != nil {
		configMapNames = parseQmPodForConfigMapNames(qmPod)
	} else {
		configMapNames = parseQueueManagerCRForConfigMapNames(dynamicClient, flags)
	}

	// fetch configmap details from configmap names
	var configMaps []corev1.ConfigMap

	for _, configMapName := range configMapNames {

		configMap, err := configmap.GetConfigMapDetailsByName(client, configMapName, flags.QueueManagerNamespace)
		if err != nil {
			logger.Info(fmt.Sprintf("Error fetching config-map %s in %s namespace: %v", configMapName, flags.QueueManagerNamespace, err))
			continue
		}
		configMaps = append(configMaps, *configMap)
	}

	if len(configMaps) != 0 {

		// write configmap to their respective yamls
		fileNameFormat := "%s.yaml"
		if err := configmap.WriteConfigMapYamlsToFile(configMaps, configMapDirectory, fileNameFormat, logger); err != nil {
			logger.Error(err.Error())
		}
	}

	if fileCount, err := utils.GetFileCountInDirectory(configMapDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("ConfigMap details: %s: Total Files: %d", configMapDirectory, fileCount))
	}

	return nil
}

func parseQmPodForConfigMapNames(qmPod *corev1.Pod) []string {

	var configMapNames []string

	for _, volume := range qmPod.Spec.Volumes {

		if volume.ConfigMap != nil {
			configMapNames = append(configMapNames, volume.ConfigMap.Name)
		}

	}

	return slices.Compact(configMapNames)

}

func parseQueueManagerCRForConfigMapNames(dynamicClient dynamic.Interface, flags utils.MustGatherFlags) []string {

	var configMapNames []string

	if flags.QueueManagerName == "" {
		return nil
	}

	queueManagerMap, err := cr.GetQueueManagerCrDetailsByName(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		return nil
	}

	configMapsAdded := make(map[string]struct{})

	addConfigMapName := func(name string) {
		if name == "" {
			return
		}

		if _, ok := configMapsAdded[name]; ok {
			return
		}

		configMapsAdded[name] = struct{}{}
		configMapNames = append(configMapNames, name)
	}

	spec := queueManagerMap["spec"].(map[string]interface{})
	if spec == nil {
		return nil
	}

	qmSpec := spec["queueManager"].(map[string]interface{})
	if qmSpec == nil {
		return nil
	}

	// spec.queueManager.mqsc
	if mqscList, ok := qmSpec["mqsc"].([]interface{}); ok {

		for _, item := range mqscList {

			if m, ok := item.(map[string]interface{}); ok {
				if cm, ok := m["configMap"].(map[string]interface{}); ok {
					if name, _ := cm["name"].(string); name != "" {
						addConfigMapName(name)
					}
				}
			}

		}

	}

	// spec.queueManager.ini
	if iniList, ok := qmSpec["ini"].([]interface{}); ok {

		for _, item := range iniList {

			if m, ok := item.(map[string]interface{}); ok {
				if cm, ok := m["configMap"].(map[string]interface{}); ok {
					if name, _ := cm["name"].(string); name != "" {
						addConfigMapName(name)
					}
				}
			}

		}

	}

	// spec.queueManager.files
	if fileList, ok := qmSpec["files"].([]interface{}); ok {

		for _, item := range fileList {

			if m, ok := item.(map[string]interface{}); ok {
				if cm, ok := m["configMap"].(map[string]interface{}); ok {
					if name, _ := cm["name"].(string); name != "" {
						addConfigMapName(name)
					}
				}
			}

		}

	}

	qmWeb := spec["web"].(map[string]interface{})
	if qmWeb == nil {
		return configMapNames
	}

	// spec.web.manualConfig

	if manualConfig, ok := qmWeb["manualConfig"].(map[string]interface{}); ok {
		if cm, ok := manualConfig["configMap"].(map[string]interface{}); ok {
			if name, _ := cm["name"].(string); name != "" {
				addConfigMapName(name)
			}
		}
	}

	return configMapNames

}
