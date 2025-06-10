package namespace

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckIfNamespacesExist(client kubernetes.Interface, namespaceNames ...string) (string, error) {
	for _, namespaceName := range namespaceNames {
		_, err := client.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
		if err != nil {
			return namespaceName, err
		}
	}
	return "", nil
}
