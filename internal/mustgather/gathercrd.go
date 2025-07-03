package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/crd"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/dynamic"
)

func gatherCrdToFiles(dynamicClient dynamic.Interface, flags utils.MustGatherFlags) error {

	// get QueueManager details by QueueManager name
	queueManagerDetailsMap, err := crd.GetQueueManagerCrdDetailsByName(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error retrieving queue-manager %s in the namespace %s: %v", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	// check if the QueueManager with the provided name exists in the provided namespace
	if queueManagerDetailsMap == nil {
		return fmt.Errorf("error no QueueManager with metadata.name as %s, found in the namespace %s", flags.QueueManagerName, flags.QueueManagerNamespace)
	}

	//write the QueueManager details to its yaml
	fileNameFormat := "%s.yaml"
	if err := crd.WriteQueueManagerCrdYamlToFiles(queueManagerDetailsMap, fileNameFormat, flags.OutputDir, flags.QueueManagerName); err != nil {
		return err
	}

	// get the IntegrationKeycloakClient details by OwnershipReferences
	integrationKeycloakClientDetails, err := crd.GetIntegrationKeycloakClientDetailsByOwnerReferences(dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		// if requested resource not found then just continue
		fmt.Printf("integration-keycloak-client %s in the namespace %s: %v\n", flags.QueueManagerName, flags.QueueManagerNamespace, err)
	}

	if integrationKeycloakClientDetails != nil {

		// write the IntegrationKecloakClient details to its yaml
		integrationkeycloakClientFileNameFormat := "%s-integration-keycloak-client.yaml"
		if err := crd.WriteIntegrationKeycloakClientCrdYamlToFiles(integrationKeycloakClientDetails, integrationkeycloakClientFileNameFormat, flags.OutputDir); err != nil {
			return err
		}
	}

	return nil

}
