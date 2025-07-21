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
package csv

import (
	"context"
	"strings"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetOperatorCSVBySelector retrieves all operator csv's in a given namespace that match the provided label selector.
// Parameters:
//   - dynamicClient: the Kubernetes dynamic client used to interact with the cluster.
//   - selector: the label selector used to filter the csv.
//   - namespace: the namespace in which to search for the csv.
func GetOperatorCSVBySelector(dynamicClient dynamic.Interface, selector, namespace string) (*unstructured.UnstructuredList, error) {

	csvGVR := schema.GroupVersionResource{
		Group:    utils.OperatorGroup,
		Version:  utils.OperatorVersion,
		Resource: utils.CSVResource,
	}

	unstructuredObject, err := dynamicClient.Resource(csvGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	return unstructuredObject, nil

}

// GetOperatorCSVByNamePrefix retrieves all operator csv's in a given namespace that has operatorNamePrefix.
// Parameters:
//   - dynamicClient: the Kubernetes dynamic client used to interact with the cluster.
//   - operatorNamePrefix: the operator name prefix used to filter the csv.
//   - namespace: the namespace in which to search for the csv.
func GetOperatorCSVByNamePrefix(dynamicClient dynamic.Interface, operatorNamePrefix, namespace string) ([]unstructured.Unstructured, error) {

	csvGVR := schema.GroupVersionResource{
		Group:    utils.OperatorGroup,
		Version:  utils.OperatorVersion,
		Resource: utils.CSVResource,
	}

	unstructuredObject, err := dynamicClient.Resource(csvGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var filteredList []unstructured.Unstructured

	for _, item := range unstructuredObject.Items {
		if strings.HasPrefix(item.GetName(), operatorNamePrefix) {
			filteredList = append(filteredList, item)
		}
	}

	return filteredList, nil

}
