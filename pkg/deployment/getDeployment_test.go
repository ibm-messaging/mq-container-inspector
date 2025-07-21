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
package deployment

import (
	"context"
	"fmt"
	"testing"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/test"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	selector  = "app.kubernetes.io/name=ibm-mq,control-plane=controller-manager"
	namespace = "testing"
)

func TestGetDeploymentsBySelector(t *testing.T) {

	expectedDeploymentCount, deploymentList, err := getExpectedMqOperatorDeploymentCountFromFakeCoreClient()
	if err != nil {
		t.Errorf("%v", err)
	}

	var runtimeObjects []runtime.Object
	for index := range deploymentList.Items {
		runtimeObjects = append(runtimeObjects, deploymentList.Items[index].DeepCopy())
	}

	fakeClient := fake.NewSimpleClientset(runtimeObjects...)

	result, err := GetDeploymentsBySelector(fakeClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching mq-operator deployment by selector: %v", err)
	}

	if got := len(result); got != expectedDeploymentCount {
		t.Errorf("got %d deployments, but expected %d deployments", got, expectedDeploymentCount)
	}

}

func getExpectedMqOperatorDeploymentCountFromFakeCoreClient() (int, appsv1.DeploymentList, error) {

	var expectedDeploymentList appsv1.DeploymentList

	coreClient, err := test.NewFakeCoreClientBySelector(selector, namespace)
	if err != nil {
		return 0, expectedDeploymentList, fmt.Errorf("error creating a new fake core client: %v", err)
	}

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return 0, expectedDeploymentList, fmt.Errorf("error while parsing label selector %s: %v", selector, err)
	}

	err = coreClient.List(context.TODO(), &expectedDeploymentList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	})
	if err != nil {
		return 0, expectedDeploymentList, fmt.Errorf("error while fetching mq-operator deployment from fake client %v", err)
	}

	return len(expectedDeploymentList.Items), expectedDeploymentList, nil

}
