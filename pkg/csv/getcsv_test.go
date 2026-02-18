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
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/test"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
