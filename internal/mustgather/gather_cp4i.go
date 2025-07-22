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

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/csv"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/dynamic"
)

func gatherCp4IOperatorCSVToFiles(dynamicClient dynamic.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

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
		logger.Info(fmt.Sprintf("could not find %s operator in %s namespace: %v\n", utils.CommonServicesOperatorPrefix, flags.QueueManagerNamespace, err))
	}

	if commonServiceOperatorCSVList != nil {
		if err := csv.WriteCSVYamlsToFile(commonServiceOperatorCSVList, fileNameFormat, cp4iDirectory); err != nil {
			logger.Error(err.Error())
		}
	}

	// get the "ibm-integration-platform-navigator" csv details
	pnOperatorCSVList, err := csv.GetOperatorCSVByNamePrefix(dynamicClient, utils.CP4iOperatorPrefix, flags.QueueManagerNamespace)
	if err != nil {
		// continue if not found
		logger.Info(fmt.Sprintf("could not find %s operator in %s namespace: %v\n", utils.CP4iOperatorPrefix, flags.QueueManagerNamespace, err))
	}

	if pnOperatorCSVList != nil {
		if err := csv.WriteCSVYamlsToFile(pnOperatorCSVList, fileNameFormat, cp4iDirectory); err != nil {
			logger.Error(err.Error())
		}
	}

	if fileCount, err := utils.GetFileCountInDirectory(cp4iDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("CP4I details: %s: Total Files: %d", cp4iDirectory, fileCount))
	}

	return nil

}
