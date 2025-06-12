package crd

import (
	"context"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetQueueManagerCrdDetailsByName retrieves the details of the QueueManager CRD with the specified name in the given namespace.
// Parameters:
//   - dynamicClient:    the Kubernetes dynamic client used to fetch CRD resources.
//   - queueManagerName: the name of the QueueManager CRD to retrieve.
//   - namespace:        the namespace in which to look up the QueueManager CRD.
func GetQueueManagerCrdDetailsByName(dynamicClient dynamic.Interface, queueManagerName, namespace string) (map[string]interface{}, error) {

	queueManagerGVR := schema.GroupVersionResource{
		Group:    utils.QmgrGroup,
		Version:  utils.QmgrVersion,
		Resource: utils.QmgrResource,
	}

	unstructuredObject, err := dynamicClient.Resource(queueManagerGVR).Namespace(namespace).Get(context.TODO(), queueManagerName, metav1.GetOptions{})

	return unstructuredObject.Object, err

}
