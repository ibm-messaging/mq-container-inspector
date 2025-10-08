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

package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/ingress"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/service"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
)

func gatherIngressToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	ingressDirectory := filepath.Join(flags.OutputDir, "ingress")
	if !utils.CheckIfDirectoryExist(ingressDirectory) {
		if err := utils.CreateDirectory(ingressDirectory, 0775); err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	var serviceList []corev1.Service
	var err error

	if flags.QueueManagerName != "" {

		labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

		serviceList, err = service.GetServiceDetailsBySelector(coreClient, labelSelector, flags.QueueManagerNamespace)
		if err != nil {
			logger.Error(fmt.Sprintf("error fetching services by selector: %v", err))
			return nil
		}
	} else if flags.PodName != "" {

		serviceList, err = service.GetServiceDetailsByPodName(coreClient, flags.PodName, flags.QueueManagerNamespace)
		if err != nil {
			logger.Error(fmt.Sprintf("error fetching services by pod-name: %v", err))
			return nil
		}

	}

	ingressList, err := getMatchingIngressFromService(coreClient, serviceList, flags, logger)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	// write the ingress in their respective yaml files
	ingressFileNameFormat := "%s-ingress.yaml"
	if err := ingress.WriteIngressYamlsToFile(ingressList, ingressFileNameFormat, ingressDirectory); err != nil {
		logger.Error(err.Error())
		return nil
	}

	if fileCount, err := utils.GetFileCountInDirectory(ingressDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Ingress Details: %s: Total Files: %d", ingressDirectory, fileCount))
	}

	return nil
}

func getMatchingIngressFromService(coreClient kubernetes.Interface, serviceList []corev1.Service, flags utils.MustGatherFlags, logger *slog.Logger) ([]networkingv1.Ingress, error) {
	// fetch all the service names
	serviceNames := make(map[string]struct{})
	for _, service := range serviceList {
		serviceNames[service.ObjectMeta.Name] = struct{}{}
	}

	// fetch all ingresses in the namespace
	ingressList, err := ingress.ListAllIngressInNamespace(coreClient, flags.QueueManagerNamespace)
	if err != nil {
		logger.Warn(fmt.Sprintf("error listing all ingress in %s namespace: %v", flags.QueueManagerNamespace, err))
		return nil, nil
	}

	var matchingIngress []networkingv1.Ingress

	for _, ingress := range ingressList {
		found := false

		// check DefaultBackend
		if db := ingress.Spec.DefaultBackend; db != nil && db.Service != nil {
			if _, ok := serviceNames[db.Service.Name]; ok {
				found = true
			}
		}

		// check spec.rules.http.paths
		if !found {
			for _, rule := range ingress.Spec.Rules {
				if rule.HTTP == nil {
					continue
				}

				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service == nil {
						continue
					}

					if _, ok := serviceNames[path.Backend.Service.Name]; ok {
						found = true
						break
					}
				}

				if found {
					break
				}
			}
		}

		if found {
			matchingIngress = append(matchingIngress, ingress)
		}

	}

	return matchingIngress, nil
}
