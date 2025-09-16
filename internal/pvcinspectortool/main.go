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

package pvcinspectortool

import (
	"fmt"
	"log/slog"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pvcinspector"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/tarzip"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/validations"
	"k8s.io/client-go/rest"
)

func PVCInspectorTool(cfg *rest.Config, flags utils.PVCInspectorFlags) error {

	// initialize logger
	logFile, err := utils.InitializeLogFile(flags.OutputDir, utils.PVCInspectorLogFileName)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Printf("error closing logFile %v", err)
		}
	}(logFile)

	handler := slog.NewTextHandler(logFile, nil)
	logger := slog.New(handler)

	// initialize the kube clients
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	dynamicClient, err := kubeclient.BuildKubernetesDynamicClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building dynamic client from config: %v", err)
	}

	// check if the provided queue manager namespace exists on the cluster
	if err := validations.ValidateNamespace(coreClient, flags.QueueManagerNamespace); err != nil {
		return err
	}

	// validate the --qm-name flag
	qmPod, err := validations.ValidateQueueManagerName(coreClient, dynamicClient, flags.QueueManagerName, flags.QueueManagerNamespace)
	if err != nil {
		return err
	}

	// validate the --pod-name flag
	if pod, err := validations.ValidatePodName(coreClient, flags.PodName, flags.QueueManagerNamespace); err != nil {
		return err
	} else if pod != nil {
		if labelValue, exists := pod.ObjectMeta.Labels["app.kubernetes.io/instance"]; exists {
			flags.QueueManagerName = labelValue
			flags.PodName = ""
		}
		qmPod = pod
	}

	logger.Info("---- Starting PVC-Inspector tool ----")

	pvcPods, err := pvcinspector.SetupPVCPods(coreClient, flags, qmPod, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error setting up PVC pods: %v", err))
		return err
	}

	// verify pvc-pods exist, will be nil with no error when dry-run is enabled
	if pvcPods != nil {
		logger.Info(fmt.Sprintf("%d pvc-inspector pods created in %s namespace", len(pvcPods), flags.QueueManagerNamespace))

		if flags.Cleanup {
			logger.Info("PVC-pods cleanup has been enabled")
			logger.Info(fmt.Sprintf("Starting cleanup for the %d pvc-inspector pods", len(pvcPods)))
			pvcinspector.DeletePVCPods(coreClient, pvcPods, flags.QueueManagerNamespace, logger)
			logger.Info("Cleanup process completed")
		}
	}

	if !flags.NoTar {
		logger.Info("tar-zip flag enabled, compressing the collected pvc-data")
		fmt.Println("compressing the collected pvc-inspector data")

		if err := tarzip.TarZipFolder(flags.OutputDir); err != nil {
			logger.Error(fmt.Sprintf("Error tar zipping the collected pvc-inspector details at path %s: %v", flags.OutputDir, err))
			return fmt.Errorf("error tar zipping the collected pvc-inspector details at path %s: %v", flags.OutputDir, err)
		}
	}

	logger.Info("---- PVC-Inspector tool run completed ----")

	fmt.Printf("PVC-Inpector tool logs can be found at: %s\n", utils.GetLogFilePath(flags.OutputDir, utils.PVCInspectorLogFileName))

	return nil

}
