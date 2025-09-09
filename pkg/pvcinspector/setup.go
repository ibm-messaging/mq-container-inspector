package pvcinspector

import (
	"context"
	"fmt"
	"log/slog"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func SetupPVCPods(coreClient kubernetes.Interface, flags utils.PVCInspectorFlags, qmPod *corev1.Pod, logger *slog.Logger) ([]corev1.Pod, error) {

	var podList []corev1.Pod

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

	// for each pod spin-up a corresponding pvc pod
	for _, pod := range podList {

		worker := pod.Spec.NodeName

		var persistentVolumeClaimNames []string

		// get the pod persistent volume claim names
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				persistentVolumeClaimNames = append(persistentVolumeClaimNames, volume.PersistentVolumeClaim.ClaimName)
			}
		}

		// create the pvc-pod skeleton structure
		pvcPod := createPVCPodStructure(pod, worker, persistentVolumeClaimNames)

		// create the pvc pod
		pvcPod, err := coreClient.CoreV1().Pods(pvcPod.ObjectMeta.Namespace).Create(context.TODO(), pvcPod, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}

		pvcPods = append(pvcPods, *pvcPod)

		logger.Info(fmt.Sprintf("Creating PVC pod %s for %s pod in %s namespace", pvcPod.ObjectMeta.Name, pod.ObjectMeta.Name, pvcPod.ObjectMeta.Namespace))

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

func createPVCPodStructure(pod corev1.Pod, worker string, persistentVolumeClaimNames []string) *corev1.Pod {

	var pvcPodVolumeMounts []corev1.VolumeMount
	var pvcPodVolumes []corev1.Volume

	for index, pvcName := range persistentVolumeClaimNames {
		volumeMount := corev1.VolumeMount{
			Name:      fmt.Sprintf("pvc%d-mount", index),
			MountPath: fmt.Sprintf("/%s", pvcName),
		}
		pvcPodVolumeMounts = append(pvcPodVolumeMounts, volumeMount)

		volume := corev1.Volume{
			Name: fmt.Sprintf("pvc%d-mount", index),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		}
		pvcPodVolumes = append(pvcPodVolumes, volume)
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
										Key:      "kubernetes.io/hostname",
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
					Name: utils.PVCInspectorContainer,
					// TODO: replace with the mq-pod image
					Image:        "registry.access.redhat.com/ubi9:latest",
					Command:      utils.GetPVCInspectorContainerCommand(),
					VolumeMounts: pvcPodVolumeMounts,
				},
			},
			Volumes: pvcPodVolumes,
		},
	}
}
