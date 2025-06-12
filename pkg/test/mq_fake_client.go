package test

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
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

	// register QueueManager CRD
	queueManagerGV := schema.GroupVersion{
		Group:   utils.QmgrGroup,
		Version: utils.QmgrVersion,
	}
	scheme.AddKnownTypes(queueManagerGV,
		&unstructured.Unstructured{},
		&unstructured.UnstructuredList{})
	metav1.AddToGroupVersion(scheme, queueManagerGV)

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

func newFakeQueueManagerCrdBySelector(selector, namespace string) (*unstructured.Unstructured, error) {

	// fetch the qm-instance name from the selector
	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	queueManagerCrd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", utils.QmgrGroup, utils.QmgrVersion),
			"kind":       utils.KindQueueManager,
			"metadata": map[string]interface{}{
				"name":      queueManagerName,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"license": map[string]interface{}{
					"accept":  true,
					"license": "L-NUUP-23NH8Y",
					"use":     "Production",
				},
				"queueManager": map[string]interface{}{
					"availability": map[string]interface{}{
						"type": "NativeHA",
					},
					"storage": map[string]interface{}{
						"queueManager": map[string]interface{}{
							"type": "persistent-claim",
						},
					},
				},
				"version": "9.4.3.0-r1",
				"web": map[string]interface{}{
					"enabled": false,
				},
			},
		},
	}
	queueManagerCrd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   utils.QmgrGroup,
		Version: utils.QmgrVersion,
		Kind:    utils.KindQueueManager,
	})

	return queueManagerCrd, nil
}

func NewFakeCoreClientBySelector(selector, namespace string) (controllerruntimeclient.Client, error) {

	scheme := buildNewScheme()

	podObjs, err := newFakePodsBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	// register pods
	var runtimeObjects []runtime.Object
	for _, obj := range podObjs {
		runtimeObjects = append(runtimeObjects, obj.(runtime.Object))
	}

	return controllerruntimefake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(runtimeObjects...).Build(), nil

}

func NewFakeDynamicClientBySelector(selector, namespace string) (*dynamicfake.FakeDynamicClient, error) {

	scheme := buildNewScheme()

	queueManagerCrdObj, err := newFakeQueueManagerCrdBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	return dynamicfake.NewSimpleDynamicClient(scheme, queueManagerCrdObj), nil

}
