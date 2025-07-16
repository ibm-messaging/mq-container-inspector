package mustgather

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/cr"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func gatherCRsToFiles(dynamicClient dynamic.Interface, flags utils.MustGatherFlags) error {

	// create crd directory to store crd files
	crDirectory := filepath.Join(flags.OutputDir, "crs")
	directoryExist := utils.CheckIfDirectoryExist(crDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(crDirectory, 0775); err != nil {
			return err
		}
	}

	// get QueueManager details by QueueManager name
	queueManagerDetailsMap, err := cr.GetQueueManagerCrDetailsByName(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		fmt.Printf("error retrieving queue manager %s in the namespace %s: %v\n", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	// check if the QueueManager with the provided name exists in the provided namespace
	if queueManagerDetailsMap == nil {
		// fetch the names of all the queue managers in the queue manager namespace
		fmt.Printf("Listing queue managers in the namespace: %s\n", flags.QueueManagerNamespace)
		queueManagerList, err := cr.ListQueueManagersInNamespace(dynamicClient, flags.QueueManagerNamespace)
		if errors.IsNotFound(err) || len(queueManagerList) == 0 {
			return fmt.Errorf("no queue managers found in the namespace: %s, Please ensure the namespace provided is the namespace where the queue manager is deployed", flags.QueueManagerNamespace)
		} else if err != nil {
			return fmt.Errorf("error retrieving queue managers in the namespace %s: %v", flags.QueueManagerNamespace, err)
		}
		printQueueManagerDetails(queueManagerList)
		return fmt.Errorf("please re-run the must gather tool with a valid queue manager metadata.name and namespace")

	}

	//write the QueueManager details to its yaml
	fileNameFormat := "%s.yaml"
	if err := cr.WriteQueueManagerCrdYamlToFiles(queueManagerDetailsMap, fileNameFormat, crDirectory, flags.QueueManagerName); err != nil {
		return err
	}

	// get the IntegrationKeycloakClient details by OwnershipReferences
	integrationKeycloakClientDetails, err := cr.GetIntegrationKeycloakClientDetailsByOwnerReferences(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		// if requested resource not found then just continue
		fmt.Printf("integration-keycloak-client %s in the namespace %s: %v\n", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	if integrationKeycloakClientDetails != nil {

		// write the IntegrationKecloakClient details to its yaml
		integrationkeycloakClientFileNameFormat := "%s-integration-keycloak-client.yaml"
		if err := cr.WriteIntegrationKeycloakClientCrdYamlToFiles(integrationKeycloakClientDetails, integrationkeycloakClientFileNameFormat, crDirectory); err != nil {
			return err
		}
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
