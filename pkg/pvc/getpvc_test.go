package pvc

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	selector  = "app.kubernetes.io/instance=QM1"
	namespace = "testing"
)

func TestGetPVCDetailsBySelector(t *testing.T) {

	expectedPvcCount, pvcList, err := getExpectedPvcCountFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	var runtimeObjects []runtime.Object
	for index := range pvcList.Items {
		runtimeObjects = append(runtimeObjects, pvcList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetPVCDetailsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching pvc details by selector: %v", err)
	}

	t.Errorf("%#v\n\n\n%#v", pvcList, result)

	if got := len(result); got != expectedPvcCount {
		t.Errorf("got %d pvc's, but expected %d pvc's", got, expectedPvcCount)
	}

}

func getExpectedPvcCountFromFakeCoreClient() (int, corev1.PersistentVolumeClaimList, error) {

	var expectedPvcList corev1.PersistentVolumeClaimList

	coreClient, err := test.NewFakeCoreClientBySelector(selector, namespace)
	if err != nil {
		return 0, expectedPvcList, fmt.Errorf("error creating a new fake core client: %v", err)
	}

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return 0, expectedPvcList, fmt.Errorf("error while parsing label selector %s: %v", selector, err)
	}

	err = coreClient.List(context.TODO(), &expectedPvcList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	})
	if err != nil {
		return 0, expectedPvcList, fmt.Errorf("error while fetching pvc's from fake client %v", err)
	}

	return len(expectedPvcList.Items), expectedPvcList, nil

}
