package kubeclient

import (
	"fmt"

	"k8s.io/client-go/discovery"
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

func CreateNewDiscoveryClient(cfg *rest.Config) (*discovery.DiscoveryClient, error) {

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return discoveryClient, nil

}
