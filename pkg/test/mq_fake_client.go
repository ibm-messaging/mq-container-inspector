package test

import (
	"fmt"

	routeV1 "github.com/openshift/api/route/v1"
	routefake "github.com/openshift/client-go/route/clientset/versioned/fake"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	routeV1.AddToScheme(scheme)

	// register QueueManager CRD
	queueManagerGV := schema.GroupVersion{
		Group:   utils.QmgrGroup,
		Version: utils.QmgrVersion,
	}
	scheme.AddKnownTypes(queueManagerGV,
		&unstructured.Unstructured{},
		&unstructured.UnstructuredList{})
	metav1.AddToGroupVersion(scheme, queueManagerGV)

	// register CSV
	csvGV := schema.GroupVersion{
		Group:   utils.OperatorGroup,
		Version: utils.OperatorVersion,
	}
	scheme.AddKnownTypeWithName(
		csvGV.WithKind(utils.KindCSV),
		&unstructured.Unstructured{},
	)
	scheme.AddKnownTypeWithName(
		csvGV.WithKind(utils.KindCSVList),
		&unstructured.UnstructuredList{},
	)
	metav1.AddToGroupVersion(scheme, csvGV)

	return scheme

}

func newFakePodsBySelector(selector, namespace string) ([]controllerruntimeclient.Object, error) {

	// fetch the qm-instance name from the selector
	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	var pods []controllerruntimeclient.Object
	for i := 0; i < 3; i++ {
		podName := fmt.Sprintf("%s-ibm-mq-%d", queueManagerName, i)
		ownerReferenceControllerAndDeletionRule := true

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/instance": queueManagerName,
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         utils.ApiVersionAppsV1,
						Kind:               utils.KindStatefulSet,
						Name:               fmt.Sprintf("%s-ibm-mq", queueManagerName),
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

func newFakeStatefulSetBySelector(selector, namespace string) (controllerruntimeclient.Object, error) {

	// fetch the qm-instance name from the selector
	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	statefulSetName := fmt.Sprintf("%s-ibm-mq", queueManagerName)
	ownerReferenceControllerAndDeletionRule := true
	replicas := int32(3)

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulSetName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/instance": queueManagerName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         fmt.Sprintf("%s/%s", utils.QmgrGroup, utils.QmgrVersion),
					Kind:               utils.KindQueueManager,
					Name:               queueManagerName,
					Controller:         &ownerReferenceControllerAndDeletionRule,
					BlockOwnerDeletion: &ownerReferenceControllerAndDeletionRule,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/instance": queueManagerName,
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
				},
			},
			Replicas: &replicas,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "data",
						Namespace: namespace,
						Labels: map[string]string{
							"app.kubernetes.io/instance": queueManagerName,
						},
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("2Gi"),
							},
						},
					},
				},
			},
		},
	}

	return statefulSet, nil

}

func newFakeStatefulSetRevisionBySelector(selector, namespace string) ([]controllerruntimeclient.Object, error) {

	// fetch the qm-instance name from the selector
	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	var revisions []controllerruntimeclient.Object
	for i := 1; i <= 3; i++ {
		ownerReferenceControllerAndDeletionRule := true

		controllerRevision := &appsv1.ControllerRevision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-ibm-mq-revision-%d", queueManagerName, i),
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/instance": queueManagerName,
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         utils.ApiVersionAppsV1,
						Kind:               utils.KindStatefulSet,
						Name:               fmt.Sprintf("%s-ibm-mq", queueManagerName),
						Controller:         &ownerReferenceControllerAndDeletionRule,
						BlockOwnerDeletion: &ownerReferenceControllerAndDeletionRule,
					},
				},
			},
			Revision: int64(i),
			Data: runtime.RawExtension{
				Raw: []byte("{}"),
			},
		}

		revisions = append(revisions, controllerRevision)
	}

	return revisions, nil

}

func newFakeServiceBySelector(selector, namespace string) ([]controllerruntimeclient.Object, error) {

	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	queueManagerServiceName := fmt.Sprintf("%s-ibm-mq", queueManagerName)
	queueManagerMetricsServiceName := fmt.Sprintf("%s-ibm-mq-metrics", queueManagerName)
	ownerReferenceControllerAndDeletionRule := true

	queueManagerService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      queueManagerServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/instance": queueManagerName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         fmt.Sprintf("%s/%s", utils.QmgrGroup, utils.QmgrVersion),
					Kind:               utils.KindQueueManager,
					Name:               queueManagerName,
					Controller:         &ownerReferenceControllerAndDeletionRule,
					BlockOwnerDeletion: &ownerReferenceControllerAndDeletionRule,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			IPFamilies: []corev1.IPFamily{
				corev1.IPv4Protocol,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "console-https",
					Protocol:   corev1.ProtocolTCP,
					Port:       9443,
					TargetPort: intstr.FromInt(9443),
				},
				{
					Name:       "qmgr",
					Protocol:   corev1.ProtocolTCP,
					Port:       1414,
					TargetPort: intstr.FromInt(1414),
				},
			},
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
		},
	}

	queueManagerMetricService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      queueManagerMetricsServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/instance": queueManagerName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         fmt.Sprintf("%s/%s", utils.QmgrGroup, utils.QmgrVersion),
					Kind:               utils.KindQueueManager,
					Name:               queueManagerName,
					Controller:         &ownerReferenceControllerAndDeletionRule,
					BlockOwnerDeletion: &ownerReferenceControllerAndDeletionRule,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			IPFamilies: []corev1.IPFamily{
				corev1.IPv4Protocol,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Protocol:   corev1.ProtocolTCP,
					Port:       9157,
					TargetPort: intstr.FromInt(9157),
				},
			},
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
		},
	}

	return []controllerruntimeclient.Object{queueManagerService, queueManagerMetricService}, nil

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
					"enabled": true,
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

func newFakePVCsBySelector(selector, namespace string) ([]controllerruntimeclient.Object, error) {

	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	storageClassName := "ocs-storagecluster-ceph-rbd"
	var pvcList []controllerruntimeclient.Object

	for i := 0; i < 3; i++ {

		pvcName := fmt.Sprintf("data-%s-volume-ibm-mq-%d", queueManagerName, i)

		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/instance": queueManagerName,
				},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("2Gi"),
					},
				},
				StorageClassName: &storageClassName,
			},
		}

		pvcList = append(pvcList, pvc)

	}

	return pvcList, nil

}

func newFakeRoutesBySelector(selector, namespace string) ([]controllerruntimeclient.Object, error) {

	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		return nil, err
	}

	qmRouteName := fmt.Sprintf("%s-ibm-mq-qm", queueManagerName)
	webRouteName := fmt.Sprintf("%s-ibm-mq-web", queueManagerName)
	ownerReferenceControllerAndDeletionRule := true
	routeTargetServiceWeight := int32(100)

	qmRoute := &routeV1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      qmRouteName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/instance": queueManagerName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         fmt.Sprintf("%s/%s", utils.QmgrGroup, utils.QmgrVersion),
					Kind:               utils.KindQueueManager,
					Name:               queueManagerName,
					Controller:         &ownerReferenceControllerAndDeletionRule,
					BlockOwnerDeletion: &ownerReferenceControllerAndDeletionRule,
				},
			},
		},
		Spec: routeV1.RouteSpec{
			Host: fmt.Sprintf("%s.sample-host.ibm.com", qmRouteName),
			To: routeV1.RouteTargetReference{
				Kind:   utils.KindService,
				Name:   fmt.Sprintf("%s-ibm-mq", queueManagerName),
				Weight: &routeTargetServiceWeight,
			},
			Port: &routeV1.RoutePort{
				TargetPort: intstr.FromInt(1414),
			},
			TLS: &routeV1.TLSConfig{
				Termination: routeV1.TLSTerminationPassthrough,
			},
			WildcardPolicy: routeV1.WildcardPolicyNone,
		},
	}

	webRoute := &routeV1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webRouteName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/instance": queueManagerName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         fmt.Sprintf("%s/%s", utils.QmgrGroup, utils.QmgrVersion),
					Kind:               utils.KindQueueManager,
					Name:               queueManagerName,
					Controller:         &ownerReferenceControllerAndDeletionRule,
					BlockOwnerDeletion: &ownerReferenceControllerAndDeletionRule,
				},
			},
		},
		Spec: routeV1.RouteSpec{
			Host: fmt.Sprintf("%s-sample-host.ibm.mq", webRouteName),
			To: routeV1.RouteTargetReference{
				Kind:   utils.KindService,
				Name:   fmt.Sprintf("%s-ibm-mq", queueManagerName),
				Weight: &routeTargetServiceWeight,
			},
			Port: &routeV1.RoutePort{
				TargetPort: intstr.FromInt(9443),
			},
			TLS: &routeV1.TLSConfig{
				Termination: routeV1.TLSTerminationPassthrough,
			},
			WildcardPolicy: routeV1.WildcardPolicyNone,
		},
	}

	return []controllerruntimeclient.Object{qmRoute, webRoute}, nil

}

func newFakeMqOperatorDeploymentBySelector(selector, namespace string) controllerruntimeclient.Object {

	replicas := int32(1)
	revisionHistoryLimit := int32(1)
	progressDeadlineSeconds := int32(600)
	terminationGracePeriodSeconds := int64(10)
	runAsNonRoot := true

	mqOperatorDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ibm-mq-operator",
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "ibm-mq",
				"control-plane":          "controller-manager",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:                &replicas,
			RevisionHistoryLimit:    &revisionHistoryLimit,
			ProgressDeadlineSeconds: &progressDeadlineSeconds,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyAlways,
					SchedulerName: "default-scheduler",
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"amd64",
													"ppc64le",
													"s390x",
												},
											},
											{
												Key:      "kubernetes.io/os",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"linux",
												},
											},
										},
									},
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
					},
					Containers: []corev1.Container{
						{
							Name:            "manager",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Image:           "na.artifactory.swg-devops.com/hyc-mq-container-team-docker-local/asiro/ibm-mq-operator:latest",
							Args: []string{
								"--leader-elect",
								"--health-probe-bind-address=:8081",
							},
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}
	return mqOperatorDeployment
}

func newFakeMqOperatorPodBySelector(selector, namespace string) controllerruntimeclient.Object {

	terminationGracePeriodSeconds := int64(10)

	mqOperatorPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ibm-mq-operator-c57f9c7b6-c84kj",
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "ibm-mq",
				"control-plane":          "controller-manager",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "kubernetes.io/arch",
										Operator: corev1.NodeSelectorOpIn,
										Values: []string{
											"amd64",
											"ppc64le",
											"s390x",
										},
									},
									{
										Key:      "kubernetes.io/os",
										Operator: corev1.NodeSelectorOpIn,
										Values: []string{
											"linux",
										},
									},
								},
							},
						},
					},
				},
			},
			TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
			Containers: []corev1.Container{
				{
					Name:            "manager",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Image:           "na.artifactory.swg-devops.com/hyc-mq-container-team-docker-local/asiro/ibm-mq-operator:latest",
					Args: []string{
						"--leader-elect",
						"--health-probe-bind-address=:8081",
					},
				},
			},
		},
	}

	return mqOperatorPod

}

func newFakeMQOperatorCSVBySelector(selector, namespace string) *unstructured.Unstructured {

	csv := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", utils.OperatorGroup, utils.OperatorVersion),
			"kind":       utils.KindCSV,
			"metadata": map[string]interface{}{
				"name":      "ibm-mq.v3.6.0",
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       "ibm-mq",
					"app.kubernetes.io/managed-by": "olm",
				},
			},
			"spec": map[string]interface{}{
				"customresourcedefinitions": []interface{}{
					map[string]interface{}{
						"displayName": "Queue Manager",
						"kind":        utils.KindQueueManager,
						"name":        "queuemanagers.mq.ibm.com",
						"version":     utils.QmgrVersion,
					},
				},
				"displayName": "IBM MQ",
				"provider": map[string]interface{}{
					"name": "IBM",
				},
				"maturity": "stable",
				"version":  "3.6.0",
			},
		},
	}
	csv.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   utils.OperatorGroup,
		Version: utils.OperatorVersion,
		Kind:    utils.KindCSV,
	})

	return csv

}

func NewFakeCoreClientBySelector(selector, namespace string) (controllerruntimeclient.Client, error) {

	scheme := buildNewScheme()

	podObjs, err := newFakePodsBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	statefulSetObj, err := newFakeStatefulSetBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	statefulSetRevisionObjs, err := newFakeStatefulSetRevisionBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	serviceObjs, err := newFakeServiceBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	pvcObjs, err := newFakePVCsBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	mqOperatorDeploymentObj := newFakeMqOperatorDeploymentBySelector(selector, namespace)
	mqOperatorPodObj := newFakeMqOperatorPodBySelector(selector, namespace)

	// register pods
	var runtimeObjects []runtime.Object
	for _, obj := range podObjs {
		runtimeObjects = append(runtimeObjects, obj.(runtime.Object))
	}

	// register statefulset
	runtimeObjects = append(runtimeObjects, statefulSetObj.(runtime.Object))

	// register statefulset revisions
	for _, obj := range statefulSetRevisionObjs {
		runtimeObjects = append(runtimeObjects, obj.(runtime.Object))
	}

	//register pvc's
	for _, pvc := range pvcObjs {
		runtimeObjects = append(runtimeObjects, pvc.(runtime.Object))
	}

	// register service
	for _, obj := range serviceObjs {
		runtimeObjects = append(runtimeObjects, obj.(runtime.Object))
	}

	// register mq-operator deployment and pod
	runtimeObjects = append(runtimeObjects, mqOperatorDeploymentObj.(runtime.Object), mqOperatorPodObj.(runtime.Object))

	return controllerruntimefake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(runtimeObjects...).Build(), nil

}

func NewFakeDynamicClientBySelector(selector, namespace string) (*dynamicfake.FakeDynamicClient, error) {

	scheme := buildNewScheme()

	// get the QueueManager object
	queueManagerCrdObj, err := newFakeQueueManagerCrdBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	// get the CSV object
	csvObj := newFakeMQOperatorCSVBySelector(selector, namespace)

	return dynamicfake.NewSimpleDynamicClient(scheme, queueManagerCrdObj, csvObj), nil

}

func NewFakeRouteClientBySelector(selector, namespace string) (*routefake.Clientset, error) {

	routeObjs, err := newFakeRoutesBySelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	// register routes
	var runtimeObjects []runtime.Object
	for _, routeObj := range routeObjs {
		runtimeObjects = append(runtimeObjects, routeObj)
	}

	return routefake.NewSimpleClientset(runtimeObjects...), nil

}
