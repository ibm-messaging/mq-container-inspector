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
package statefulset

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/test"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"

	controllerrevisions "github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/controller_revisions"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	selector  = "app.kubernetes.io/instance=QM1"
	namespace = "testing"
)

func TestGetStatefulSetDetailsBySelector(t *testing.T) {

	expectedStatefulSetCount, statefulSetList, err := getExpectedStatefulSetListFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	var runtimeObjects []runtime.Object
	for index := range statefulSetList.Items {
		runtimeObjects = append(runtimeObjects, statefulSetList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetStatefulSetDetailsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching StatefulSet by selector: %v", err)
	}

	if got := len(result); got != expectedStatefulSetCount {
		t.Errorf("got %d StatefulSet, but expected %d StatefulSet", got, expectedStatefulSetCount)
	}

}

func TestGetStatefulSetRevisionsBySelector(t *testing.T) {

	expectedStatefulSetRevisionCount, statefulSetRevisionList, err := getExpectedStatefulSetRevisionListFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	var runtimeObjects []runtime.Object
	for index := range statefulSetRevisionList.Items {
		runtimeObjects = append(runtimeObjects, statefulSetRevisionList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := controllerrevisions.GetControllerRevisionsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching StatefulSet revisions by selector: %v", err)
	}

	if got := len(result); got != expectedStatefulSetRevisionCount {
		t.Errorf("got %d StatefulSet revisions, expected %d StatefulSet revisions", got, expectedStatefulSetRevisionCount)
	}

}

func TestGetStatefulSetEventsBySelector(t *testing.T) {

	expectedStatefulSetCount, statefulSetList, err := getExpectedStatefulSetListFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	expecteStatefulSetNames := getExpectedStatefulSetNames(statefulSetList)

	var runtimeObjects []runtime.Object
	for index := range statefulSetList.Items {
		runtimeObjects = append(runtimeObjects, statefulSetList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetStatefulSetEventsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching StatefulSet events by selector: %v", err)
	}

	if got := len(result); got != expectedStatefulSetCount {
		t.Errorf("got %d StatefulSet, but expected %d StatefulSet", got, expectedStatefulSetCount)
	}

	// check if the got and expected names match
	for _, expectedStatefulSetName := range expecteStatefulSetNames {
		if _, ok := result[expectedStatefulSetName]; !ok {
			t.Errorf("error expected StatefulSet %q to be a key in result map, but it was missing", expectedStatefulSetName)
		}
	}
}

func getExpectedStatefulSetRevisionListFromFakeCoreClient() (int, appsv1.ControllerRevisionList, error) {

	var expectedStatefulSetRevisionList appsv1.ControllerRevisionList

	coreClient, err := test.NewFakeCoreClientBySelector(selector, namespace)
	if err != nil {
		return 0, expectedStatefulSetRevisionList, fmt.Errorf("error creating a new fake core client: %v", err)
	}

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return 0, expectedStatefulSetRevisionList, fmt.Errorf("error while parsing label selector %s: %v", selector, err)
	}

	err = coreClient.List(context.TODO(), &expectedStatefulSetRevisionList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	})
	if err != nil {
		return 0, expectedStatefulSetRevisionList, fmt.Errorf("error while fetching StatefulSet revisions from fake client %v", err)
	}

	return len(expectedStatefulSetRevisionList.Items), expectedStatefulSetRevisionList, nil

}

func getExpectedStatefulSetListFromFakeCoreClient() (int, appsv1.StatefulSetList, error) {

	var expectedStatefulSetList appsv1.StatefulSetList

	coreClient, err := test.NewFakeCoreClientBySelector(selector, namespace)
	if err != nil {
		return 0, expectedStatefulSetList, fmt.Errorf("error creating a new fake core client: %v", err)
	}

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return 0, expectedStatefulSetList, fmt.Errorf("error while parsing label selector %s: %v", selector, err)
	}

	err = coreClient.List(context.TODO(), &expectedStatefulSetList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	})
	if err != nil {
		return 0, expectedStatefulSetList, fmt.Errorf("error while fetching StatefulSet from fake client %v", err)
	}

	return len(expectedStatefulSetList.Items), expectedStatefulSetList, nil
}

func getExpectedStatefulSetNames(statefulSetList appsv1.StatefulSetList) []string {

	var statefulSetNames []string
	for _, statefulSet := range statefulSetList.Items {
		statefulSetNames = append(statefulSetNames, statefulSet.Name)
	}

	return statefulSetNames

}
