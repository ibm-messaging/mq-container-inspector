package mustgather

import (
	"fmt"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/csv"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/dynamic"
)

func gatherCp4IOperatorCSVToFiles(dynamicClient dynamic.Interface, flags utils.MustGatherFlags) error {

	// create cp4i directory to store cp4i files
	cp4iDirectory := filepath.Join(flags.OutputDir, "cp4i")
	directoryExists := utils.CheckIfDirectoryExist(cp4iDirectory)
	if !directoryExists {
		if err := utils.CreateDirectory(cp4iDirectory, 0775); err != nil {
			return err
		}
	}

	fileNameFormat := "%s-csv.yaml"

	// get the "ibm-common-service-operator" csv details
	commonServiceOperatorCSVList, err := csv.GetOperatorCSVByNamePrefix(dynamicClient, utils.CommonServicesOperatorPrefix, flags.QueueManagerNamespace)
	if err != nil {
		// continue if not found
		fmt.Printf("could not find %s operator in %s namespace: %v", utils.CommonServicesOperatorPrefix, flags.QueueManagerNamespace, err)
	}

	if commonServiceOperatorCSVList != nil && len(commonServiceOperatorCSVList) > 0 {
		if err := csv.WriteCSVYamlsToFile(commonServiceOperatorCSVList, fileNameFormat, cp4iDirectory); err != nil {
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
		if err := csv.WriteCSVYamlsToFile(pnOperatorCSVList, fileNameFormat, cp4iDirectory); err != nil {
			return err
		}
	}

	return nil

}
