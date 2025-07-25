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
