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
package cr

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetQueueManagerCrDetailsByName retrieves the details of the QueueManager CR with the specified name in the given namespace.
// Parameters:
//   - dynamicClient:    the Kubernetes dynamic client used to fetch CR resources.
//   - queueManagerName: the name of the QueueManager CR to retrieve.
//   - namespace:        the namespace in which to look up the QueueManager CR.
func GetQueueManagerCrDetailsByName(dynamicClient dynamic.Interface, queueManagerName, namespace string) (map[string]interface{}, error) {

	queueManagerGVR := schema.GroupVersionResource{
		Group:    utils.QmgrGroup,
		Version:  utils.QmgrVersion,
		Resource: utils.QmgrResource,
	}

	unstructuredObject, err := dynamicClient.Resource(queueManagerGVR).Namespace(namespace).Get(context.TODO(), queueManagerName, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return unstructuredObject.Object, err

}

// GetIntegrationKeycloakClientDetailsByOwnerReferences retrieves the details of the IntegrationKeycloakClient CR with the specified owner-references in the given namespace.
// Returns (nil,nil) if there are no IntegrationKeycloakClients found.
// Parameters:
//   - dynamicClient:    the Kubernetes dynamic client used to fetch CR resources.
//   - queueManagerName: the name of the QueueManager CR which will be in owner-reference of the IntegrationKeycloakClient CR.
//   - namespace:        the namespace in which to look up the IntegrationKeycloakClient CR.
func GetIntegrationKeycloakClientDetailsByOwnerReferences(dynamicClient dynamic.Interface, queueManagerName, namespace string) ([]*unstructured.Unstructured, error) {

	integrationKeycloakClientGVR := schema.GroupVersionResource{
		Group:    utils.IntegrationKeycloakClientGroup,
		Version:  utils.IntegrationKeycloakClientVersion,
		Resource: utils.IntegrationKeycloakClientResource,
	}

	unstructuredList, err := dynamicClient.Resource(integrationKeycloakClientGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
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

	return itegrationKeycloakClientList, nil

}

// ListQueueManagersInNamespace retrieves all the QueueManagers in the given namespace.
// Parameters:
//   - dynamicClient:    the Kubernetes dynamic client used to fetch CR resources.
//   - namespace:        the namespace in which to look up the QueueManager CR.
func ListQueueManagersInNamespace(dynamicClient dynamic.Interface, namespace string) ([]unstructured.Unstructured, error) {

	queueManagerGVR := schema.GroupVersionResource{
		Group:    utils.QmgrGroup,
		Version:  utils.QmgrVersion,
		Resource: utils.QmgrResource,
	}

	unstructuredObject, err := dynamicClient.Resource(queueManagerGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return unstructuredObject.Items, err

}
