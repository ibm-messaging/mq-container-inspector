/*
© Copyright IBM Corporation 2026

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

package utils

import (
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
)

func FilterDeploymentListByAnnotation(deploymentList []appsv1.Deployment, annotationKey string, annotationValue string) []appsv1.Deployment {
	var deployments []appsv1.Deployment

	for _, deployment := range deploymentList {

		annotations := deployment.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			deployments = append(deployments, deployment)
		}

	}

	return deployments
}

func FilterReplicasetListByAnnotation(replicasetList []appsv1.ReplicaSet, annotationKey string, annotationValue string) []appsv1.ReplicaSet {

	var replicasets []appsv1.ReplicaSet

	for _, replicaset := range replicasetList {

		annotations := replicaset.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			replicasets = append(replicasets, replicaset)
		}
	}

	return replicasets
}

func FilterServiceListByAnnotation(servicesList []corev1.Service, annotationKey string, annotationValue string) []corev1.Service {

	var services []corev1.Service

	for _, service := range servicesList {
		annotations := service.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			services = append(services, service)
		}
	}

	return services
}

func FilterRouteListByAnnotation(routeList []routev1.Route, annotationKey string, annotationValue string) []routev1.Route {

	var routes []routev1.Route

	for _, route := range routeList {
		annotations := route.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			routes = append(routes, route)
		}
	}

	return routes
}

func FilterNetworkPolicyListByAnnotation(networkPolicyList []networkv1.NetworkPolicy, annotationKey, annotationValue string) []networkv1.NetworkPolicy {

	var filteredNetworkPolicyList []networkv1.NetworkPolicy

	for _, networkPolicy := range networkPolicyList {

		annotations := networkPolicy.Annotations

		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			filteredNetworkPolicyList = append(filteredNetworkPolicyList, networkPolicy)
		}
	}

	return filteredNetworkPolicyList

}

func FilterServiceAccountListByAnnotation(serviceAccountList []corev1.ServiceAccount, annotationKey string, annotationValue string) []corev1.ServiceAccount {

	var serviceAccounts []corev1.ServiceAccount

	for _, serviceAccount := range serviceAccountList {
		annotations := serviceAccount.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			serviceAccounts = append(serviceAccounts, serviceAccount)
		}
	}

	return serviceAccounts
}

func FilterConfigMapListByAnnotation(configMapList []corev1.ConfigMap, annotationKey, annotationValue string) []corev1.ConfigMap {

	var configMaps []corev1.ConfigMap

	for _, configMap := range configMapList {
		annotations := configMap.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			configMaps = append(configMaps, configMap)
		}
	}

	return configMaps
}
