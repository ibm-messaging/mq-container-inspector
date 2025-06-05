package test

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	controllerruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	controllerruntimefake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func buildNewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	clientgoscheme.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	storagev1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)

	return scheme

}

func newFakePodsBySelector(selector, namespace string) ([]controllerruntimeclient.Object, error) {

	// fetch the qm-instance name from the selector
	qmName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	var pods []controllerruntimeclient.Object
	for i := 0; i < 3; i++ {
		podName := fmt.Sprintf("%s-ibm-mq-%d", qmName, i)
		ownerReferenceControllerAndDeletionRule := true

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/instance": qmName,
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         utils.ApiVersionAppsV1,
						Kind:               utils.KindStatefulSet,
						Name:               fmt.Sprintf("%s-ibm-mq", qmName),
						Controller:         &ownerReferenceControllerAndDeletionRule,
						BlockOwnerDeletion: &ownerReferenceControllerAndDeletionRule,
					},
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "qmgr",
						Image: "cp.icr.io/cp/ibm-mqadvanced-server",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 1414,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								ContainerPort: 9157,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: fmt.Sprintf("data-%s", podName),
							},
						},
					},
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}

		pods = append(pods, pod)
	}

	return pods, nil

}

func NewFakeCoreClientBySelector(selector, namespace string) (controllerruntimeclient.Client, error) {

	scheme := buildNewScheme()

	podObjs, err := newFakePodsBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	var runtimePodObjects []runtime.Object
	for _, obj := range podObjs {
		runtimePodObjects = append(runtimePodObjects, obj.(runtime.Object))
	}

	return controllerruntimefake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(runtimePodObjects...).Build(), nil

}
