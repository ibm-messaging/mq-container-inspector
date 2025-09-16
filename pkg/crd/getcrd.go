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

package crd

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetCRDDetailsByName retrieves the CRD details with the specified name.
// Parameters:
//   - dynamicClient:    the Kubernetes dynamic client used to fetch CR resources.
//   - crdName:          the name of the CRD to retrieve.
func GetCRDDetailsByName(dynamicClient dynamic.Interface, crdName string) (map[string]interface{}, error) {

	crdGVR := schema.GroupVersionResource{
		Group:    utils.CRDGroup,
		Version:  utils.ApiVersionV1,
		Resource: utils.CRDResource,
	}

	unstructuredObject, err := dynamicClient.Resource(crdGVR).Get(context.TODO(), crdName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return unstructuredObject.Object, err

}
