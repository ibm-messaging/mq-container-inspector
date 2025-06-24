package namespace

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DoesNamespacesExist(client kubernetes.Interface, namespace string) (bool, error) {

	_, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
