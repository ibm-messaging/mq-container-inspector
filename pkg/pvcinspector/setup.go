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
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
	}

	// for each pod spin-up a corresponding pvc pod
	for _, pod := range podList {

		worker := pod.Spec.NodeName

		// create the pvc-pod skeleton structure
		pvcPod := createPVCPodStructure(pod, worker)
		var err error

		if !flags.DryRun {
			// create the pvc pod
			pvcPod, err = coreClient.CoreV1().Pods(pvcPod.ObjectMeta.Namespace).Create(context.TODO(), pvcPod, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}

			pvcPods = append(pvcPods, *pvcPod)

			logger.Info(fmt.Sprintf("Creating PVC pod %s for %s pod in %s namespace", pvcPod.ObjectMeta.Name, pod.ObjectMeta.Name, pvcPod.ObjectMeta.Namespace))

		}

		// collect the pvc-pod yaml
		if err := collectPVCPodYaml(pvcPod, pvcPodYamlDir); err != nil {
			logger.Error(fmt.Sprintf("Error collecting pvc-pod yaml: %v", err))
		}

	}

	return pvcPods, nil

}

func DeletePVCPods(coreClient kubernetes.Interface, pvcPods []corev1.Pod, namespace string, logger *slog.Logger) {

	if len(pvcPods) > 0 {

		for _, pvcPod := range pvcPods {
			err := coreClient.CoreV1().Pods(namespace).Delete(context.TODO(), pvcPod.ObjectMeta.Name, metav1.DeleteOptions{})
			if err != nil {
				logger.Error(fmt.Sprintf("Error deleting %s pvc-pod: %v", pvcPod.ObjectMeta.Name, err))
			} else {
				logger.Info(fmt.Sprintf("Successfully deleted %s pvc-pod", pvcPod.ObjectMeta.Name))
			}
		}

	}

}

func createPVCPodStructure(pod corev1.Pod, worker string) *corev1.Pod {

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

	return generatePVCPodSkeleton(pod, worker, pvcPodVolumeMounts, pvcPodVolumes)
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
