package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/crd"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/dynamic"
)

func gatherCRDsToFiles(dynamicClient dynamic.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	qmgrCRDName := fmt.Sprintf("%s.%s", utils.QmgrResource, utils.QmgrGroup)

	queueManagerCRDDetails, err := crd.GetCRDDetailsByName(dynamicClient, qmgrCRDName)
	if err == nil && queueManagerCRDDetails != nil {

		// create QueueManager directory if it does not exists
		queueManagerDirectory := filepath.Join(flags.OutputDir, "queue-managers")
		if directoryExists := utils.CheckIfDirectoryExist(queueManagerDirectory); !directoryExists {
			if err := utils.CreateDirectory(queueManagerDirectory, 0775); err != nil {
				return err
			}
		}

		crdFileNameFormat := "queue-manager-crd.yaml"
		if err := crd.WriteCRDYamlsToFile(queueManagerCRDDetails, crdFileNameFormat, queueManagerDirectory); err != nil {
			logger.Error(err.Error())
		}

		if fileCount, err := utils.GetFileCountInDirectory(queueManagerDirectory); err != nil {
			logger.Error(err.Error())
		} else {
			logger.Info(fmt.Sprintf("Queue-Manager details: %s: Total Files: %d", queueManagerDirectory, fileCount))
		}
	} else {
		logger.Info(fmt.Sprintf("queue-manager CRD not found: %v", err))
	}

	return nil
}
