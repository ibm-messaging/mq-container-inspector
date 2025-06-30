package crd

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/test"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

const (
	selector  = "app.kubernetes.io/instance=QM1"
	namespace = "testing"
)

func TestGetQueueManagerCrdDetailsByName(t *testing.T) {

	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		t.Errorf("%v", err)
	}

	dynamicFakeClient, err := test.NewFakeDynamicClientBySelector(selector, namespace)
	if err != nil {
		t.Errorf("error creating a new fake dynamic client: %v", err)
	}

	expectedQueueManagerName, err := getExpectedQueueManagerNameFromFakeDynamicClient(dynamicFakeClient, queueManagerName)
	if err != nil {
		t.Errorf("%v", err)
	}

	queueManagerDetails, err := GetQueueManagerCrdDetailsByName(dynamicFakeClient, queueManagerName, namespace)
	if err != nil {
		t.Errorf("error getting queue-manager %s by name: %v", queueManagerName, err)
	}

	gotQueueManagerName := getQueueManagerName(queueManagerDetails)
	if gotQueueManagerName != expectedQueueManagerName {
		t.Errorf("got %s queue-manager, expected %s queue-manager", gotQueueManagerName, expectedQueueManagerName)
	}

}

func TestGetIntegrationKeycloakClientDetailsByOwnerReferences(t *testing.T) {

	queueManagerName, err := utils.FetchQMGRResourceNameFromSelector(selector)
	if err != nil {
		t.Errorf("%v", err)
	}

	dynamicFakeClient, err := test.NewFakeDynamicClientBySelector(selector, namespace)
	if err != nil {
		t.Errorf("error creating a new fake dynamic client: %v", err)
	}

	expectedintegrationKeycloakClientCount, err := getExpectedintegrationKeycloakClientCountFromFakeDynamicClient(dynamicFakeClient, queueManagerName)
	if err != nil {
		t.Errorf("%v", err)
	}

	result, err := GetIntegrationKeycloakClientDetailsByOwnerReferences(dynamicFakeClient, queueManagerName, namespace)
	if err != nil {
		t.Errorf("error getting integration-keycloak-client with owner queue-manager %s: %v", queueManagerName, err)
	}

	if got := len(result); got != expectedintegrationKeycloakClientCount {
		t.Errorf("got %d integration-keycloak-client, expected %d integration-keycloak-client", got, expectedintegrationKeycloakClientCount)
	}

}

func getQueueManagerName(queueManagerDetails map[string]interface{}) string {

	metadataMap := queueManagerDetails["metadata"].(map[string]interface{})
	return metadataMap["name"].(string)

}

func getExpectedQueueManagerNameFromFakeDynamicClient(dynamicFakeClient *fake.FakeDynamicClient, queueManagerName string) (string, error) {

	queueManagerGVR := schema.GroupVersionResource{
		Group:    utils.QmgrGroup,
		Version:  utils.QmgrVersion,
		Resource: utils.QmgrResource,
	}

	unstructuredObject, err := dynamicFakeClient.Resource(queueManagerGVR).Namespace(namespace).Get(context.TODO(), queueManagerName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting queue-manager %s from dynamic fake client: %v", queueManagerName, err)
	}

	return unstructuredObject.GetName(), nil

}

func getExpectedintegrationKeycloakClientCountFromFakeDynamicClient(dynamicFakeClient *fake.FakeDynamicClient, queueManagerName string) (int, error) {

	integrationKeycloakClientGVR := schema.GroupVersionResource{
		Group:    utils.IntegrationKeycloakClientGroup,
		Version:  utils.IntegrationKeycloakClientVersion,
		Resource: utils.IntegrationKeycloakClientResource,
	}

	unstructuredList, err := dynamicFakeClient.Resource(integrationKeycloakClientGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return 0, err
	}

	var itegrationKeycloakClientList []*unstructured.Unstructured

	for index := range unstructuredList.Items {
		item := &unstructuredList.Items[index]

		for _, ownerReference := range item.GetOwnerReferences() {
			if ownerReference.Kind == utils.KindQueueManager && ownerReference.Name == queueManagerName {
				itegrationKeycloakClientList = append(itegrationKeycloakClientList, item)
				break
			}
		}
	}

	return len(itegrationKeycloakClientList), nil
}
