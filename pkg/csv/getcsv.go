package csv

import (
	"context"

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
