package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/csv"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func gatherCp4IOperatorCSVToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build dynamic client from config
	dynamicClient, err := kubeclient.BuildKubernetesDynamicClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building dynamic client from config: %v", err)
	}

	fileNameFormat := "%s-csv.yaml"

	// get the "ibm-common-service-operator" csv details
	commonServiceOperatorCSVList, err := csv.GetOperatorCSVByNamePrefix(dynamicClient, utils.CommonServicesOperatorPrefix, flags.QueueManagerNamespace)
	if err != nil {
		// continue if not found
		fmt.Printf("could not find %s operator in %s namespace: %v", utils.CommonServicesOperatorPrefix, flags.QueueManagerNamespace, err)
	}

	if commonServiceOperatorCSVList != nil && len(commonServiceOperatorCSVList) > 0 {
		if err := csv.WriteCSVYamlsToFile(commonServiceOperatorCSVList, fileNameFormat, flags.OutputDir); err != nil {
			return err
		}
	}

	// get the "ibm-integration-platform-navigator" csv details
	pnOperatorCSVList, err := csv.GetOperatorCSVByNamePrefix(dynamicClient, utils.CP4iOperatorPrefix, flags.QueueManagerNamespace)
	if err != nil {
		// continue if not found
		fmt.Printf("could not find %s operator in %s namespace: %v", utils.CP4iOperatorPrefix, flags.QueueManagerNamespace, err)
	}

	if pnOperatorCSVList != nil && len(pnOperatorCSVList) > 0 {
		if err := csv.WriteCSVYamlsToFile(pnOperatorCSVList, fileNameFormat, flags.OutputDir); err != nil {
			return err
		}
	}

	return nil

}
