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
	"encoding/base64"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/cr"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
)

func gatherCRsToFiles(dynamicClient dynamic.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create queue-managers directory to store queue-managers cr files
	qmDirectory := filepath.Join(flags.OutputDir, "queue-managers")
	directoryExist := utils.CheckIfDirectoryExist(qmDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(qmDirectory, 0775); err != nil {
			return err
		}
	}

	// get QueueManager details by QueueManager name
	queueManagerDetailsMap, err := cr.GetQueueManagerCrDetailsByName(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if errors.IsNotFound(err) || queueManagerDetailsMap == nil {
		logger.Info(fmt.Sprintf("queue manager custom resource '%s' not found in the namespace '%s'", flags.QueueManagerName, flags.QueueManagerNamespace))
		return nil
	} else if err != nil {
		fmt.Printf("error retrieving queue manager %s in the namespace %s: %v\n", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	logger.Info(fmt.Sprintf("Found %s queue manager in %s namespace", flags.QueueManagerName, flags.QueueManagerNamespace))

	// check if we have lastValidConfiguration in the queue-manager cr
	if ok, lastValidConfiguration := queueManagerCrHasLastValidConfiguration(queueManagerDetailsMap); ok {
		logger.Info(fmt.Sprintf("'lastValidConfiguration' found in the status.metadata.lastValidConfiguration for %s queue-manager", flags.QueueManagerName))
		decodedBytes, err := base64.StdEncoding.DecodeString(lastValidConfiguration)
		if err != nil {
			logger.Info(fmt.Sprintf("Error decoding lastValidConfiguration: %v", err))
		}

		// write the lastValidConfiguration to its yaml file
		fileNameFormat := "%s-last-valid-config.yaml"
		if err := cr.WriteQueueManagerLastValidConfigToFiles(decodedBytes, fileNameFormat, qmDirectory, flags.QueueManagerName); err != nil {
			logger.Info(fmt.Sprintf("%s", err.Error()))
		}
	}

	//write the QueueManager details to its yaml
	fileNameFormat := "%s.yaml"
	if err := cr.WriteQueueManagerCrdYamlToFiles(queueManagerDetailsMap, fileNameFormat, qmDirectory, flags.QueueManagerName); err != nil {
		return err
	}

	// get the IntegrationKeycloakClient details by OwnershipReferences
	integrationKeycloakClientDetails, err := cr.GetIntegrationKeycloakClientDetailsByOwnerReferences(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		// if requested resource not found then just continue
		logger.Info(fmt.Sprintf("integration-keycloak-client %s in the namespace %s: %v", flags.QueueManagerName, flags.QueueManagerNamespace, err))
	}

	if integrationKeycloakClientDetails != nil {

		logger.Info(fmt.Sprintf("Found IntegrationKeycloakClient resource with %s queue manager as owner, in %s namespace", flags.QueueManagerName, flags.QueueManagerNamespace))

		// create cp4i directory to store cp4i files
		cp4iDirectory := filepath.Join(flags.OutputDir, "cp4i")
		directoryExists := utils.CheckIfDirectoryExist(cp4iDirectory)
		if !directoryExists {
			if err := utils.CreateDirectory(cp4iDirectory, 0775); err != nil {
				return err
			}
		}

		// write the IntegrationKecloakClient details to its yaml
		integrationkeycloakClientFileNameFormat := "%s-integration-keycloak-client.yaml"
		if err := cr.WriteIntegrationKeycloakClientCrdYamlToFiles(integrationKeycloakClientDetails, integrationkeycloakClientFileNameFormat, cp4iDirectory); err != nil {
			logger.Error(err.Error())
		}
	}

	if fileCount, err := utils.GetFileCountInDirectory(qmDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Queue-Manager details: %s: Total Files: %d", qmDirectory, fileCount))
	}

	return nil

}

func queueManagerCrHasLastValidConfiguration(queueManagerDetailsMap map[string]interface{}) (bool, string) {

	qmStatusVal, ok := queueManagerDetailsMap["status"]
	if !ok || qmStatusVal == nil {
		return false, ""
	}

	qmStatus, ok := qmStatusVal.(map[string]interface{})
	if !ok {
		return false, ""
	}

	qmMetaValue, ok := qmStatus["metadata"]
	if !ok || qmMetaValue == nil {
		return false, ""
	}

	qmStatusMetadata, ok := qmMetaValue.(map[string]interface{})
	if !ok {
		return false, ""
	}

	// Get "lastValidConfiguration" safely
	lastValidConfigvalue, ok := qmStatusMetadata["lastValidConfiguration"]
	if !ok || lastValidConfigvalue == nil {
		return false, ""
	}

	lastValidConfiguration, ok := lastValidConfigvalue.(string)
	if !ok || lastValidConfiguration == "" {
		return false, ""
	}

	return true, lastValidConfiguration

}
