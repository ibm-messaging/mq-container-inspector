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
package pods

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/test"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

const (
	selector  = "app.kubernetes.io/instance=QM1"
	namespace = "testing"
)

func TestGetPodsBySelector(t *testing.T) {

	expectedPodCount, podList, err := getExpectedPodCountFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	var runtimeObjects []runtime.Object
	for index := range podList.Items {
		runtimeObjects = append(runtimeObjects, podList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetPodsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching pods by selector: %v", err)
	}

	if got := len(result); got != expectedPodCount {
		t.Errorf("got %d pods, but expected %d pods", got, expectedPodCount)
	}
}

func TestGetPodLogsBySelector(t *testing.T) {

	expectedPodCount, podList, err := getExpectedPodCountFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedPodNames := getExpectedPodNames(podList)

	var runtimeObjects []runtime.Object
	for index := range podList.Items {
		runtimeObjects = append(runtimeObjects, podList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetPodLogsBySelector(fakeClient, selector, namespace, utils.QmgrContainer)
	if err != nil {
		t.Errorf("error fetching pods logs by selector: %v", err)
	}

	// check if the got and expected podcount's match
	if got := len(result); got != expectedPodCount {
		t.Errorf("got %d pods, but expected %d pods", got, expectedPodCount)
	}

	// check if the got and expected names match
	for _, expectedPodName := range expectedPodNames {
		if _, ok := result[expectedPodName]; !ok {
			t.Errorf("error expected pod %q to be a key in result map, but it was missing", expectedPodName)
		}
	}
}

func TestGetPodEventsBySelector(t *testing.T) {

	expectedPodCount, podList, err := getExpectedPodCountFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedPodNames := getExpectedPodNames(podList)

	var runtimeObjects []runtime.Object
	for index := range podList.Items {
		runtimeObjects = append(runtimeObjects, podList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetPodEventsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching pods events by selector: %v", err)
	}

	// check if the got and expected podcount's match
	if got := len(result); got != expectedPodCount {
		t.Errorf("got %d pods, but expected %d pods", got, expectedPodCount)
	}

	// check if the got and expected names match
	for _, expectedPodName := range expectedPodNames {
		if _, ok := result[expectedPodName]; !ok {
			t.Errorf("error expected pod %q to be a key in result map, but it was missing", expectedPodName)
		}
	}

}

func getExpectedPodCountFromFakeCoreClient() (int, corev1.PodList, error) {

	var expectedPodList corev1.PodList

	coreClient, err := test.NewFakeCoreClientBySelector(selector, namespace)
	if err != nil {
		return 0, expectedPodList, fmt.Errorf("error creating a new fake core client: %v", err)
	}

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return 0, expectedPodList, fmt.Errorf("error while parsing label selector %s: %v", selector, err)
	}

	err = coreClient.List(context.TODO(), &expectedPodList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	})
	if err != nil {
		return 0, expectedPodList, fmt.Errorf("error while fetching pods from fake client %v", err)
	}

	return len(expectedPodList.Items), expectedPodList, nil

}

func getExpectedPodNames(podList corev1.PodList) []string {

	var podNames []string
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames

}
