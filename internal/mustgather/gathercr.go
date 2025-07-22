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
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/cr"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		fmt.Printf("\nERROR: queue manager '%s' not found in the namespace '%s'\n", flags.QueueManagerName, flags.QueueManagerNamespace)
		// find all the queue managers in the queue manager namespace
		queueManagerList, err := cr.ListQueueManagersInNamespace(dynamicClient, flags.QueueManagerNamespace)
		if errors.IsNotFound(err) || len(queueManagerList) == 0 {
			fmt.Printf("No queue managers found in the namespace '%s'. Please validate your namespace is correct\n", flags.QueueManagerNamespace)
			return fmt.Errorf("queue manager '%s' not found", flags.QueueManagerName)
		} else if err != nil {
			return fmt.Errorf("error retrieving queue managers in the namespace %s: %v", flags.QueueManagerNamespace, err)
		}
		fmt.Printf("Available queue managers in namespace '%s':\n", flags.QueueManagerNamespace)
		printQueueManagerDetails(queueManagerList)
		fmt.Printf("Please re-run the must gather with a valid queue manager name\n")
		return fmt.Errorf("queue manager '%s' not found", flags.QueueManagerName)
	} else if err != nil {
		fmt.Printf("error retrieving queue manager %s in the namespace %s: %v\n", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	logger.Info(fmt.Sprintf("Found %s queue manager in %s namespace", flags.QueueManagerName, flags.QueueManagerNamespace))

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
		logger.Info(fmt.Sprintf("CR details: %s: Total Files: %d", qmDirectory, fileCount))
	}

	return nil

}

func printQueueManagerDetails(queueManagerList []unstructured.Unstructured) {

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// print header
	fmt.Fprintln(writer, "NAME\tPHASE")

	for _, queueManager := range queueManagerList {
		name := queueManager.GetName()
		phase, found, err := unstructured.NestedString(queueManager.Object, "status", "phase")
		if err != nil || !found {
			phase = "<unknown>"
		}
		fmt.Fprintf(writer, "%s\t%s\n", name, phase)
	}

	writer.Flush()

}
