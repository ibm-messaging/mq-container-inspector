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

package pvcinspector

import (
	"fmt"
	"log/slog"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/pvc"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type QMImageInfo struct {
	Image              string // fully resolved image reference, ready to use in a pod spec
	Version            string // VRMF-release version, if one could be determined; "" otherwise
	QMCRFound          bool   // whether the QueueManager CR was found during validation
	StorageIsEphemeral bool   // best-effort: true if the CR confirms ephemeral storage
}

type volumeSpec struct {
	name      string
	mountPath string
}

var (
	dataVolumeSpec          = volumeSpec{name: "data", mountPath: "/mnt/mqm"}
	persistedDataVolumeSpec = volumeSpec{name: "persisted-data", mountPath: "/mnt/mqm-data"}
	recoveryLogsVolumeSpec  = volumeSpec{name: "recovery-logs", mountPath: "/mnt/mqm-log"}
)

type pvcRef struct {
	volSpec volumeSpec
	pvcName string
}

type PodPVCMountData struct {
	Volumes      []corev1.Volume
	VolumeMounts []corev1.VolumeMount
}

func createPVCPodsByQMImage(coreClient kubernetes.Interface, flags utils.PVCInspectorFlags, qmImageInfo QMImageInfo, logger *slog.Logger) (map[string]PodPVCMountData, error) {

	// check if the qm-name flag is not empty
	if flags.QueueManagerName == "" {
		return nil, fmt.Errorf("when falling back to qm-image, the qm-name flag cannot be empty")
	}

	podPVCMountDataMap, err := getPVCMountDataByDiscovery(coreClient, flags, logger)
	if err != nil {
		return nil, err
	}

	if len(podPVCMountDataMap) == 0 {
		if qmImageInfo.QMCRFound && qmImageInfo.StorageIsEphemeral {
			return nil, fmt.Errorf("no PVCs found for QueueManager %q in namespace %s: the QueueManager CR specifies ephemeral storage", flags.QueueManagerName, flags.QueueManagerNamespace)
		}

		return nil, fmt.Errorf("no supported PVCs were found for QueueManager %q in namespace %q", flags.QueueManagerName, flags.QueueManagerNamespace)
	}

	return podPVCMountDataMap, nil

}

func buildPodPVCMountDataMap(coreClient kubernetes.Interface, pvcInspectorPodPVCRefsMap map[string][]pvcRef, namespace string, logger *slog.Logger) map[string]PodPVCMountData {

	podPVCMountDataMap := make(map[string]PodPVCMountData, len(pvcInspectorPodPVCRefsMap))

	for podName, refs := range pvcInspectorPodPVCRefsMap {

		var volumes []corev1.Volume
		var volumeMounts []corev1.VolumeMount

		for _, ref := range refs {

			pvcDetails, err := pvc.GetPVCDetailByName(coreClient, ref.pvcName, namespace)
			if err != nil {
				logger.Warn(
					"Skipping PVC because it could not be retrieved",
					"pvc", ref.pvcName,
					"namespace", namespace,
					"pvcInspectorPod", podName,
					"error", err,
				)
				continue
			}

			volumes = append(volumes, corev1.Volume{
				Name: ref.volSpec.name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcDetails.Name,
						ReadOnly:  true,
					},
				},
			})

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      ref.volSpec.name,
				MountPath: ref.volSpec.mountPath,
				ReadOnly:  true,
			})

		}

		if len(volumes) == 0 || len(volumeMounts) == 0 {
			logger.Warn(
				"Skipping PVC-inspector pod because no verified PVCs were found",
				"pvcInspectorPod", podName,
				"namespace", namespace,
			)
			continue
		}

		podPVCMountDataMap[podName] = PodPVCMountData{
			Volumes:      volumes,
			VolumeMounts: volumeMounts,
		}
	}

	return podPVCMountDataMap

}
