package kubeclient

import (
	"fmt"

	routeClient "github.com/openshift/client-go/route/clientset/versioned"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func BuildKubeConfig(kubeconfigPath string) (*rest.Config, error) {

	if kubeconfigPath == "" {
		return nil, fmt.Errorf("kubeconfigPath shouldnot be empty")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func BuildDiscoveryClientFromConfig(cfg *rest.Config) (*discovery.DiscoveryClient, error) {
	return discovery.NewDiscoveryClientForConfig(cfg)
}

func BuildKubernetesClientFromConfig(cfg *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(cfg)
}

func BuildKubernetesDynamicClientFromConfig(cfg *rest.Config) (*dynamic.DynamicClient, error) {
	return dynamic.NewForConfig(cfg)
}

func BuildRouteClientFromConfig(cfg *rest.Config) (*routeClient.Clientset, error) {
	return routeClient.NewForConfig(cfg)
}
