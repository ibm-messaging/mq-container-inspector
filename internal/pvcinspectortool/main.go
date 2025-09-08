package pvcinspectortool

import (
	"fmt"
	"log/slog"
	"os"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pvcinspector"
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
	logger.Info(fmt.Sprintf("%d pvc-inspector pods created in %s namespace", len(pvcPods), flags.QueueManagerNamespace))

	if flags.Cleanup {
		logger.Info("PVC_pods cleanup has been enabled")
		logger.Info(fmt.Sprintf("Starting cleanup for the %d pvc-inspector pods", len(pvcPods)))
		pvcinspector.DeletePVCPods(coreClient, pvcPods, flags.QueueManagerNamespace, logger)
		logger.Info("Cleanup process completed")
	}

	logger.Info("---- PVC-Inspector tool run completed ----")

	fmt.Printf("PVC-Inpector tool logs can be found at: %s\n", utils.GetLogFilePath(flags.OutputDir, utils.PVCInspectorLogFileName))

	return nil

}
