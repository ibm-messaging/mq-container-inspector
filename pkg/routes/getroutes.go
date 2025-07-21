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
package routes

import (
	"context"

	routeV1 "github.com/openshift/api/route/v1"
	routeClient "github.com/openshift/client-go/route/clientset/versioned"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetRouteDetailsBySelector retrieves all routes created by the mq-operator in a given namespace that match the provided label selector.
// Parameters:
//   - routeClient: the Route client used to interact with the cluster.
//   - selector: the label selector used to filter the routes.
//   - namespace: the namespace in which to search for the routes.
func GetRouteDetailsBySelector(routeClient routeClient.Interface, selector, namespace string) ([]routeV1.Route, error) {

	routeList, err := routeClient.RouteV1().Routes(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})

	// the route list API returns the routes with the apiVersion and kind field as empty, so setting them explicitly
	for index := range routeList.Items {
		routeList.Items[index].TypeMeta.APIVersion = utils.APIVersionRouteV1
		routeList.Items[index].TypeMeta.Kind = utils.KindRoute
	}

	return routeList.Items, err

}

// GetRouteDetailsByFieldSelector retrieves all routes to the QueueManager in a given namespace that match the provided label selector.
// Parameters:
//   - routeClient: the Route client used to interact with the cluster.
//   - fieldSelector: the field selector used to filter the routes.
//   - namespace: the namespace in which to search for the routes.
func GetRouteDetailsByFieldSelector(routeClient routeClient.Interface, fieldSelector, namespace string) ([]routeV1.Route, error) {

	routeList, err := routeClient.RouteV1().Routes(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})

	// the route list API returns the routes with the apiVersion and kind field as empty, so setting them explicitly
	for index := range routeList.Items {
		routeList.Items[index].TypeMeta.APIVersion = utils.APIVersionRouteV1
		routeList.Items[index].TypeMeta.Kind = utils.KindRoute
	}

	return routeList.Items, err

}
