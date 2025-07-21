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
