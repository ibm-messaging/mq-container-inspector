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
	"os"
	"strings"
	"time"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/tarzip"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/validations"
	"k8s.io/client-go/rest"
)

func MustGather(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build the required clients
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	dynamicClient, err := kubeclient.BuildKubernetesDynamicClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building dynamic client from config: %v", err)
	}

	routeClient, err := kubeclient.BuildRouteClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building route client: %v", err)
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

	// If OutputDir doesnot exist create the directory
	directoryExist := utils.CheckIfDirectoryExist(flags.OutputDir)
	if !directoryExist {
		if err := utils.CreateDirectory(flags.OutputDir, 0775); err != nil {
			return err
		}
	}

	// initialize logger
	logFile, err := utils.InitializeLogFile(flags.OutputDir, utils.MustGatherLogFileName)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing logFile %v", err)
		}
	}(logFile)

	handler := slog.NewTextHandler(logFile, nil)
	logger := slog.New(handler)

	logger.Info("---- Starting Must-Gather tool ----")

	mustGatherToolStartTime := time.Now()

	if flags.QueueManagerName != "" {
		// collect queue manager cr must-gathers
		logger.Info("---- Collecting queue manager details ----")
		fmt.Print("Collecting queue manager details ...")
		mustGatherStartTime := time.Now()
		err = gatherCRsToFiles(dynamicClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- queue manager details collected ----")

		// collect queue manager crd must-gathers
		logger.Info("---- Collecting queue manager CRD details ----")
		fmt.Print("Collecting queue manager CRD details ...")
		mustGatherStartTime = time.Now()
		err = gatherCRDsToFiles(dynamicClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- queue manager crd details collected ----")
	}

	// collect pods must-gathers
	logger.Info("---- Collecting pod details ----")
	fmt.Print("Collecting pod details...")
	mustGatherStartTime := time.Now()
	failedPodNames, err := gatherPodsToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
	logger.Info("---- Pod details collected ----")

	// collect the route must-gathers
	logger.Info("---- Collecting route details ----")
	fmt.Print("Collecting route details...")
	mustGatherStartTime = time.Now()
	err = gatherRoutesToFiles(cfg, coreClient, routeClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
	logger.Info("---- Route details collected ----")

	// collect the ingress must-gathers
	logger.Info("---- Collecting ingress details ----")
	fmt.Print("Collecting ingress details...")
	mustGatherStartTime = time.Now()
	err = gatherIngressToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
	logger.Info("---- Ingress details collected ----")

	// identify the pod-owner
	var podOwner string

	if qmPod == nil && flags.QueueManagerName != "" {
		// If the queuemanager pod is nil default to StatefulSet
		podOwner = utils.KindStatefulSet
	} else {
		podOwner = utils.GetPodOwner(qmPod)
	}

	// based on the pod owner collect the required must-gathers
	switch podOwner {
	case utils.KindStatefulSet:
		logger.Info(fmt.Sprintf("Found %s as the controller owner", podOwner))
		// collect StatefulSet must-gathers
		logger.Info("---- Collecting StatefulSet details ----")
		fmt.Print("Collecting StatefulSet details...")
		mustGatherStartTime = time.Now()
		err = gatherStatefulSetToFiles(coreClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- StatefulSet details collected ----")

	case utils.KindReplicaSet:
		logger.Info(fmt.Sprintf("Found %s as the controller owner", podOwner))
		// collect Deployment must-gathers
		logger.Info("---- Collecting Deployment details ----")
		fmt.Print("Collecting Deployment details...")
		mustGatherStartTime = time.Now()
		err = gatherDeploymentToFiles(coreClient, flags, qmPod, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- Deployment details collected ----")

	case utils.KindDaemonSet:
		logger.Info(fmt.Sprintf("Found %s as the as the controller owner", podOwner))
		// collect DaemonSet muust-gathers
		logger.Info("---- Collecting DaemonSet details ----")
		fmt.Print("Collecting DaemonSet details...")
		mustGatherStartTime = time.Now()
		err = gatherDaemonSetToFiles(coreClient, flags, qmPod, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- DaemonSet details collected ----")

	default:
		logger.Info(fmt.Sprintf("No owner found for pod '%s'", qmPod.ObjectMeta.Name))
	}

	// collect Service must-gathers
	logger.Info("---- Collecting service details ----")
	fmt.Print("Collecting service details...")
	mustGatherStartTime = time.Now()
	err = gatherServicesToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
	logger.Info("--- Service details collected ----")

	// collect PVC must-gathers
	logger.Info("---- Collecting PVC details ----")
	fmt.Print("Collecting PVC details...")
	mustGatherStartTime = time.Now()
	err = gatherPVCToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
	logger.Info("---- PVC details collected ----")

	if flags.QueueManagerName != "" {
		// collect MQ operator must-gathers
		logger.Info("---- Collecting mq-operator details(This may take time) ----")
		fmt.Print("Collecting mq-operator details...")
		mustGatherStartTime = time.Now()
		err = gatherMQOperatorToFiles(coreClient, dynamicClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- MQ-Operator details collected ----")

		// collect the cp4i csv details
		logger.Info("---- Collecting cp4i details ----")
		fmt.Print("Collecting cp4i details...")
		mustGatherStartTime = time.Now()
		err = gatherCp4IOperatorCSVToFiles(dynamicClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- CP4I details collected ----")
	}

	// if skip-exec is disabled then collect the web-console and runmqras logs
	if !flags.SkipExec {
		// collect web-console logs
		logger.Info("---- Collecting web-console details ----")
		fmt.Print("Collecting web-console details...")
		mustGatherStartTime = time.Now()
		err = gatherMQWebConsoleLogsToFiles(cfg, coreClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- Web-console details collected ----")

		// collect runmqras logs
		logger.Info("---- Collecting runmqras details ----")
		fmt.Print("Collecting runmqras details(This may take time) ...")
		mustGatherStartTime = time.Now()
		err = gatherRunmqrasLogToFiles(cfg, coreClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
		logger.Info("---- Runmqras details collected ----")
	}

	// delete the empty directories
	if err := utils.DeleteEmptyDirectories(flags.OutputDir, logger); err != nil {
		logger.Error(err.Error())
	}

	// if skip-tar is disabled, then tar the must-gather output
	if !flags.SkipTar {
		fmt.Print("Compressing the logs...")
		mustGatherStartTime = time.Now()

		if err := tarzip.TarZipFolder(flags.OutputDir); err != nil {
			return fmt.Errorf("error tar zipping the collected must-gather at path %s: %v", flags.OutputDir, err)
		}

		fmt.Printf("Must Gathers archived. Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
	}

	fmt.Printf("Must-Gather tool run completed... Took: %v\n", time.Since(mustGatherToolStartTime))
	logger.Info("---- Must-Gather tool run completed ----")

	fmt.Printf("Must-gather's collected, logs can be found at: %s\n\n", utils.GetLogFilePath(flags.OutputDir, utils.MustGatherLogFileName))

	// If we have found failed pods, then recommend the must-gather tool command
	if len(failedPodNames) > 0 {
		logger.Info(fmt.Sprintf("Must-Gather tool run detected %d failing pods: [%s]", len(failedPodNames), strings.Join(failedPodNames, ", ")))

		pvcInspectorCommand := fmt.Sprintf("./mq-container-inspector pvctool --pod-name %s --qm-namespace %s", failedPodNames[0], flags.QueueManagerNamespace)

		fmt.Printf(`Must-Gather tool run detected %d failing pods.

This tool is unable to gather runmqras logs for failing pods.
We recommend running the following command to create pvc-inspector pods, which will collect the failing pods PVC data:

For non-airgap: %s
For airgap:    %s
`, len(failedPodNames), pvcInspectorCommand, fmt.Sprintf("%s --dry-run", pvcInspectorCommand))

	}

	return nil

}
