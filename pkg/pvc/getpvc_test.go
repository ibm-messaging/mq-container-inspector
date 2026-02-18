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
package pvc

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/test"

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
