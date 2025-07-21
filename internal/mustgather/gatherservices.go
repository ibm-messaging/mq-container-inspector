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

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/service"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
)

func gatherServicesToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create services directory to store service files
	serviceDirectory := filepath.Join(flags.OutputDir, "services")
	directoryExist := utils.CheckIfDirectoryExist(serviceDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(serviceDirectory, 0775); err != nil {
			return err
		}
	}

	serviceLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get the service details by selector
	serviceList, err := service.GetServiceDetailsBySelector(coreClient, serviceLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching services with selector %s: %v", serviceLabelSelector, err))
	}

	// write the services to their respective yaml files
	serviceDetailsFileNameFormat := "%s-service.yaml"
	if err := service.WriteServiceYamlsToFile(serviceList, serviceDetailsFileNameFormat, serviceDirectory); err != nil {
		logger.Error(err.Error())
	}

	if fileCount, err := utils.GetFileCountInDirectory(serviceDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Service details: %s: Total Files: %d", serviceDirectory, fileCount))
	}

	return nil

}
