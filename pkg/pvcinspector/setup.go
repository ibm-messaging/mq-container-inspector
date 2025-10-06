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

package pvcinspector

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/container"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/runmqras"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Embed the custom runmqras isa file in the binary
//
//go:embed custom-isa.xml
var customISA embed.FS

func SetupPVCPods(coreClient kubernetes.Interface, flags utils.PVCInspectorFlags, qmPod *corev1.Pod, logger *slog.Logger) ([]corev1.Pod, error) {

	// check if the qmPod has persisted storage
	if !utils.CheckIfPodHasPersistedStorage(*qmPod) {
		return nil, fmt.Errorf("no PVCs found. Check that the specified resource is not using ephemeral storage")
	}

	var podList []corev1.Pod

	// check the instance type of the qmPod, and fill the podList
	if utils.GetPodInstance(qmPod) != utils.SingleInstance {
		pods, err := pods.GetMQReplicaPodsViaService(coreClient, qmPod.ObjectMeta.Name, flags.QueueManagerNamespace)
		if err != nil {
			return nil, err
		}
		podList = append(podList, pods...)
	} else {
		podList = append(podList, *qmPod)
	}

	var pvcPods []corev1.Pod

	// create pvc pods yaml directory inside the outputDir to collect the pvc-pods yamls
	pvcPodYamlDir := filepath.Join(flags.OutputDir, "pvc-pod-yamls")
	if !utils.CheckIfDirectoryExist(pvcPodYamlDir) {
		if err := utils.CreateDirectory(pvcPodYamlDir, 0775); err != nil {
			return nil, err
		}
	}

	if flags.DryRun {
		logger.Info("Dry-run enabled: pvc-inspector pods will not be created; only the pod YAMLs will be collected")
	} else {
		fmt.Println("----- Creating PVC-inspector pods -----")
	}

	// for each pod spin-up a corresponding pvc pod
	for _, pod := range podList {

		worker := pod.Spec.NodeName

		var err error

		// create the pvc-pod skeleton structure
		pvcPod, err := createPVCPodStructure(coreClient, pod, worker, flags, logger)
		if err != nil {
			return nil, err
		}
		var createdPod *corev1.Pod

		if !flags.DryRun {
			// create the pvc pod
			createdPod, err = coreClient.CoreV1().Pods(pvcPod.ObjectMeta.Namespace).Create(context.TODO(), pvcPod, metav1.CreateOptions{})
			if err != nil {
				if errors.IsAlreadyExists(err) {
					fmt.Printf("Pod %v already exists, using existing pod for pvctool\n", pvcPod.ObjectMeta.Name)
					logger.Info(fmt.Sprintf("Pod %v already exists, using existing pod for pvctool\n", pvcPod.ObjectMeta.Name))
					createdPod, err = pods.GetPodByName(coreClient, pvcPod.ObjectMeta.Name, pvcPod.ObjectMeta.Namespace)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				// watch for the pod creation and when pod has been created/failed then log the result
				if err := pods.SetupPodWatcher(coreClient, createdPod, flags.QueueManagerNamespace, utils.PodCreationWatcher, logger); err != nil {
					logger.Error(fmt.Sprintf("Error while watching pvc-inspector pod creation: %v", err))
				}
				logger.Info(fmt.Sprintf("Creating PVC pod %s for %s pod in %s namespace", createdPod.ObjectMeta.Name, pod.ObjectMeta.Name, createdPod.ObjectMeta.Namespace))
			}
			pvcPods = append(pvcPods, *createdPod)
		}

		// collect the pvc-pod yaml
		if err := collectPVCPodYaml(createdPod, pvcPodYamlDir); err != nil {
			logger.Error(fmt.Sprintf("Error collecting pvc-pod yaml: %v", err))
		}

	}
	fmt.Println("----- PVC-inspector creation process completed -----")

	return pvcPods, nil

}

func DeletePVCPods(coreClient kubernetes.Interface, pvcPods []corev1.Pod, flags utils.PVCInspectorFlags, logger *slog.Logger) {

	// Delete the pvc-pods
	if len(pvcPods) > 0 {

		fmt.Println("----- Cleaning-up the PVC-inspector pods(This may take some time) -----")

		for _, pvcPod := range pvcPods {
			err := coreClient.CoreV1().Pods(flags.QueueManagerNamespace).Delete(context.TODO(), pvcPod.ObjectMeta.Name, metav1.DeleteOptions{})
			if err != nil {
				logger.Error(fmt.Sprintf("Error deleting %s pvc-pod: %v", pvcPod.ObjectMeta.Name, err))
				fmt.Printf("Error deleting %s pvc-pod: %v\n", pvcPod.ObjectMeta.Name, err)
				return
			}

			// setup the watcher for pvc-pod cleanup
			if err := pods.SetupPodWatcher(coreClient, &pvcPod, flags.QueueManagerNamespace, utils.PodDeletionWatcher, logger); err != nil {
				logger.Error(fmt.Sprintf("Error while watching pvc-inspector pod deletion: %v", err))
			}

		}

		fmt.Println("----- PVC-inspector pods clean-up completed -----")

	}

	if flags.Runmqras {

		// Delete the custom ConfigMap created to be passed as input-file in the runmqras command
		logger.Info(fmt.Sprintf("Deleting %s ConfigMap in %s namespace", utils.CustomISAConfigMap, flags.QueueManagerNamespace))

		configMap, err := coreClient.CoreV1().ConfigMaps(flags.QueueManagerNamespace).Get(context.TODO(), utils.CustomISAConfigMap, metav1.GetOptions{})
		if err != nil {
			logger.Error(fmt.Sprintf("Error checking %s configmap in %s namespace", utils.CustomISAConfigMap, flags.QueueManagerNamespace))
			return
		}

		// delete the ConfigMap only if it is managed by mq-inspector
		if managedByMQInspector(configMap) {

			if err := coreClient.CoreV1().ConfigMaps(flags.QueueManagerNamespace).Delete(context.TODO(), utils.CustomISAConfigMap, metav1.DeleteOptions{}); err != nil {
				logger.Error(fmt.Sprintf("Error deleting %s ConfigMap in %s namespace: %v", utils.CustomISAConfigMap, flags.QueueManagerNamespace, err))
			}
			logger.Info(fmt.Sprintf("Successfully deleted %s ConfigMap in %s namespace", utils.CustomISAConfigMap, flags.QueueManagerNamespace))
		}

	}

}

func ExecuteRunmqras(cfg *rest.Config, client kubernetes.Interface, pods []corev1.Pod, flags utils.PVCInspectorFlags, logger *slog.Logger) error {

	// run the runmqras with the -inputfile flag
	runmqrasCopyConfigs, err := runmqras.ExecRunmqrasBySelector(cfg, client, pods, flags.QueueManagerNamespace, utils.PVCInspectorContainer, logger, true)
	if err != nil {
		// continue in case of error, and do not return an error so the tool doesn't stop
		logger.Error(fmt.Sprintf("Error while executing runmqras for the pvc-inspector tool: %v", err))
		return nil
	}

	for _, runmqrasCopyConfig := range runmqrasCopyConfigs {
		if err := container.CopyPathToFile(runmqrasCopyConfig, flags.OutputDir, 10); err != nil {
			logger.Error(err.Error())
		}
	}

	return nil

}

func createPVCPodStructure(client kubernetes.Interface, pod corev1.Pod, worker string, flags utils.PVCInspectorFlags, logger *slog.Logger) (*corev1.Pod, error) {

	var pvcPodVolumeMounts []corev1.VolumeMount
	var pvcPodVolumes []corev1.Volume
	volumeNameMap := make(map[string]struct{})

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcPodVolumes = append(pvcPodVolumes, volume)
			volumeNameMap[volume.Name] = struct{}{}
		}
	}

	for _, volumeMount := range pod.Spec.Containers[0].VolumeMounts {
		if _, exists := volumeNameMap[volumeMount.Name]; exists {
			pvcPodVolumeMounts = append(pvcPodVolumeMounts, volumeMount)
		}
	}

	pvcPod := generatePVCPodSkeleton(pod, worker, pvcPodVolumeMounts, pvcPodVolumes)

	// create a ConfigMap with the file-data, if configmap already exists then update the ConfigMap
	if err := createConfigMap(client, flags.QueueManagerNamespace, logger); err != nil {
		return nil, err
	}

	// update the volumes and volume-mount to mount the custom-isa.xml file
	pvcPod = mountFileOnPod(pvcPod)

	return pvcPod, nil
}

// generatePVCPodSkeleton generates the skeleton for the pvc-pod
func generatePVCPodSkeleton(pod corev1.Pod, worker string, pvcPodVolumeMounts []corev1.VolumeMount, pvcPodVolumes []corev1.Volume) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       utils.KindPod,
			APIVersion: utils.ApiVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pvc-inspector-%s", pod.ObjectMeta.Name),
			Namespace: pod.ObjectMeta.Namespace,
			Labels: map[string]string{
				"tool": "pvc-inspector-tool",
			},
		},
		Spec: corev1.PodSpec{
			// adding node affinity to handle RWO volumes
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      utils.NodeAffinityHostNameKey,
										Operator: corev1.NodeSelectorOpIn,
										Values: []string{
											worker,
										},
									},
								},
							},
						},
					},
				},
			},
			SecurityContext: pod.Spec.SecurityContext,
			Containers: []corev1.Container{
				{
					Name:         utils.PVCInspectorContainer,
					Image:        pod.Spec.Containers[0].Image,
					Command:      utils.GetPVCInspectorContainerCommand(),
					VolumeMounts: pvcPodVolumeMounts,
					Env: []corev1.EnvVar{
						{
							Name:  "LICENSE",
							Value: "accept",
						},
					},
					SecurityContext: pod.Spec.Containers[0].SecurityContext,
				},
			},
			Volumes: pvcPodVolumes,
		},
	}
}

func collectPVCPodYaml(pod *corev1.Pod, outputDir string) error {
	fileNameFormat := "%s.yaml"
	return pods.WritePodYamlsToFile([]corev1.Pod{*pod}, fileNameFormat, outputDir)

}

func createConfigMap(client kubernetes.Interface, namespace string, logger *slog.Logger) error {

	configMapExists := true

	var configMap *corev1.ConfigMap
	var err error

	configMap, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), utils.CustomISAConfigMap, metav1.GetOptions{})
	if errors.IsNotFound(err) || configMap == nil {
		configMapExists = false
	} else if err != nil {
		logger.Error(fmt.Sprintf("Error checking %s configmap in %s namespace", utils.CustomISAConfigMap, namespace))
		return err
	}

	// check if the ConfigMap is not empty and has the owner label
	if configMap.Data != nil && !managedByMQInspector(configMap) {
		// the ConfigMap with the same name already exists, without the managed-by label
		logger.Error(fmt.Sprintf("Error creating ConfigMap in namespace %s: ConfigMap %s already exists", namespace, utils.CustomISAConfigMap))
		return fmt.Errorf("error creating ConfigMap in namespace %s: ConfigMap %s already exists. Please delete the %s ConfigMap and re-run the pvc-inspector tool", namespace, utils.CustomISAConfigMap, utils.CustomISAConfigMap)
	}

	if configMapExists {
		// update the existing ConfigMap
		logger.Info(fmt.Sprintf("ConfigMap %s found in %s namespace, updating the ConfigMap", configMap.ObjectMeta.Name, namespace))

		configMap, err = getConfigMap(configMap, namespace)
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating ConfigMap in namespace %s: %v", namespace, err))
			return nil
		}

		configMap, err = client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if err != nil {
			logger.Error(fmt.Sprintf("Error updating %s config-name in %s namespace: %v", configMap.ObjectMeta.Name, namespace, err))
			return nil
		}

	} else {
		// create a new ConfigMap
		logger.Info(fmt.Sprintf("ConfigMap %s not found in %s namespace, creating the ConfigMap", utils.CustomISAConfigMap, namespace))

		configMap, err = getConfigMap(configMap, namespace)
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating ConfigMap in namespace %s: %v", namespace, err))
			return fmt.Errorf("error creating ConfigMap in namespace %s: %v", namespace, err)
		}

		configMap, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating %s config-name in %s namespace: %v", configMap.ObjectMeta.Name, namespace, err))
			return fmt.Errorf("error creating %s config-name in %s namespace: %v", configMap.ObjectMeta.Name, namespace, err)
		}

	}

	return nil

}

func getConfigMap(configMap *corev1.ConfigMap, namespace string) (*corev1.ConfigMap, error) {

	fileData, err := loadCustomISA()
	if err != nil {
		return nil, err
	}

	if configMap.Data != nil {
		configMap.Data = map[string]string{
			utils.CustomISAFileName: fileData,
		}
	} else {
		configMap = &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: utils.ApiVersionV1,
				Kind:       utils.KindConfigMap,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      utils.CustomISAConfigMap,
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "mq-container-inspector",
				},
			},
			Data: map[string]string{
				utils.CustomISAFileName: fileData,
			},
		}
	}

	return configMap, nil

}

func loadCustomISA() (string, error) {
	data, err := customISA.ReadFile(utils.CustomISAFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded custom-isa.xml: %v", err)
	}
	return string(data), nil
}

func mountFileOnPod(pod *corev1.Pod) *corev1.Pod {
	// append the ConfigMap in the volume-mount and volume
	volumeMounts := pod.Spec.Containers[0].VolumeMounts
	volumes := pod.Spec.Volumes

	customISAConfingVolumeMount := &corev1.VolumeMount{
		Name:      "custom-isa",
		MountPath: fmt.Sprintf("/run/%s", utils.CustomISAFileName),
		SubPath:   utils.CustomISAFileName,
	}

	customISAConfigVolume := &corev1.Volume{
		Name: "custom-isa",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: utils.CustomISAConfigMap,
				},
			},
		},
	}

	volumeMounts = append(volumeMounts, *customISAConfingVolumeMount)
	volumes = append(volumes, *customISAConfigVolume)

	pod.Spec.Containers[0].VolumeMounts = volumeMounts
	pod.Spec.Volumes = volumes

	return pod
}

func managedByMQInspector(configMap *corev1.ConfigMap) bool {

	key, val, _ := strings.Cut(utils.MQInspectorManagedLabel, ":")

	if configMap == nil || configMap.Labels == nil {
		return false
	}

	return configMap.Labels[key] == val

}
