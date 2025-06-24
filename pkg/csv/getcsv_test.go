package csv

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/test"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

const (
	selector  = "app.kubernetes.io/name=ibm-mq,app.kubernetes.io/managed-by=olm"
	namespace = "testing"
)

func TestGetOperatorCSVBySelector(t *testing.T) {

	dynamicfakeClient, err := test.NewFakeDynamicClientBySelector(selector, namespace)
	if err != nil {
		t.Errorf("error creating a new fake dynamic client: %v", err)
	}

	expectedCSVCount, err := getExpectedCSVCountFromFakeDynamicClient(dynamicfakeClient)
	if err != nil {
		t.Errorf("%v", err)
	}

	csvDetailsList, err := GetOperatorCSVBySelector(dynamicfakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error getting csv details by selector %s: %v", selector, err)
	}

	if got := len(csvDetailsList.Items); got != expectedCSVCount {
		t.Errorf("got %d csv's, but expected %d csv's", got, expectedCSVCount)
	}

}

func getExpectedCSVCountFromFakeDynamicClient(dynamicfakeClient *fake.FakeDynamicClient) (int, error) {

	csvGVR := schema.GroupVersionResource{
		Group:    utils.OperatorGroup,
		Version:  utils.OperatorVersion,
		Resource: utils.CSVResource,
	}

	unstructuredList, err := dynamicfakeClient.Resource(csvGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return 0, fmt.Errorf("error getting csv from dynamic fake client: %v", err)
	}

	return len(unstructuredList.Items), nil

}
