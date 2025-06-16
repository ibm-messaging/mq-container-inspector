package routes

import (
	"context"
	"fmt"
	"testing"

	routeV1 "github.com/openshift/api/route/v1"
	"github.com/openshift/client-go/route/clientset/versioned/fake"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	selector      = "app.kubernetes.io/instance=QM1"
	fieldSelector = "spec.to.name=QM1-ibm-mq"
	namespace     = "testing"
)

func TestGetRouteDetailsBySelector(t *testing.T) {

	fakeRouteClient, err := test.NewFakeRouteClientBySelector(selector, namespace)
	if err != nil {
		t.Errorf("error creating new fake route client: %v", err)
	}

	expectedRoutes, err := getExpectedRouteListFromFakeRouteClient(fakeRouteClient)
	if err != nil {
		t.Errorf("%v", err)
	}

	gotRoutes, err := GetRouteDetailsBySelector(fakeRouteClient, selector, namespace)
	if err != nil {
		t.Errorf("error fetching routes by selector: %v", err)
	}

	if len(gotRoutes) != len(expectedRoutes) {
		t.Errorf("got %d routes, but expected %d routes", len(gotRoutes), len(expectedRoutes))
	}

}

func TestGetRouteDetailsByFieldSelector(t *testing.T) {

	fakeRouteClient, err := test.NewFakeRouteClientBySelector(selector, namespace)
	if err != nil {
		t.Errorf("error creating new fake route client: %v", err)
	}

	expectedRoutes, err := getExpectedRouteListFromFakeRouteClient(fakeRouteClient)
	if err != nil {
		t.Errorf("%v", err)
	}

	gotRoutes, err := GetRouteDetailsByFieldSelector(fakeRouteClient, fieldSelector, namespace)
	if err != nil {
		t.Errorf("error fetching routes by field selector: %v", err)
	}

	if len(gotRoutes) != len(expectedRoutes) {
		t.Errorf("got %d routes, but expected %d routes", len(gotRoutes), len(expectedRoutes))
	}

}

func getExpectedRouteListFromFakeRouteClient(fakeRouteClient *fake.Clientset) ([]routeV1.Route, error) {

	routeList, err := fakeRouteClient.RouteV1().Routes(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting route list from fake client: %v", err)
	}

	return routeList.Items, nil

}
