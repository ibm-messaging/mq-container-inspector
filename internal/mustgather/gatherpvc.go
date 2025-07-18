package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pvc"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
)

func gatherPVCToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create pvc directory to store pvc files
	pvcDirectory := filepath.Join(flags.OutputDir, "pvcs")
	directoryExist := utils.CheckIfDirectoryExist(pvcDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(pvcDirectory, 0775); err != nil {
			return err
		}
	}

	pvcLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get pvc's by selector
	pvcList, err := pvc.GetPVCDetailsBySelector(coreClient, pvcLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching pvc's with selector %s: %v", pvcLabelSelector, err))
	}

	// write the pvc's to their respective yaml files
	pvcYamlFileNameFormat := "%s-pvc.yaml"
	if err := pvc.WritePVCYamlsToFile(pvcList, pvcYamlFileNameFormat, pvcDirectory); err != nil {
		logger.Error(err.Error())
	}

	if fileCount, err := utils.GetFileCountInDirectory(pvcDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("PVC Details: %s: Total Files: %d", pvcDirectory, fileCount))
	}

	return nil

}
