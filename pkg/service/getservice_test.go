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
package service

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/test"
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

func TestGetServiceDetailsBySelector(t *testing.T) {

	expectedServiceCount, serviceList, err := getExpectedServiceCountFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	var runtimeObjects []runtime.Object
	for index := range serviceList.Items {
		runtimeObjects = append(runtimeObjects, serviceList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetServiceDetailsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching service details by selector: %v", err)
	}

	if got := len(result); got != expectedServiceCount {
		t.Errorf("got %d services, but expected %d services", got, expectedServiceCount)
	}

}

func getExpectedServiceCountFromFakeCoreClient() (int, corev1.ServiceList, error) {

	var expectedServiceList corev1.ServiceList

	coreClient, err := test.NewFakeCoreClientBySelector(selector, namespace)
	if err != nil {
		return 0, expectedServiceList, fmt.Errorf("error creating a new fake core client: %v", err)
	}

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return 0, expectedServiceList, fmt.Errorf("error while parsing label selector %s: %v", selector, err)
	}

	err = coreClient.List(context.TODO(), &expectedServiceList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	})
	if err != nil {
		return 0, expectedServiceList, fmt.Errorf("error while fetching services from fake client %v", err)
	}

	return len(expectedServiceList.Items), expectedServiceList, nil

}
