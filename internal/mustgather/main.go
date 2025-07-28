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
	"time"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/namespace"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/tarzip"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func MustGather(cfg *rest.Config, flags utils.MustGatherFlags) error {

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
	namespaceExists, err := namespace.DoesNamespacesExist(coreClient, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while checking if the namespace %v exists: %v", flags.QueueManagerNamespace, err)
	} else if !namespaceExists {
		return fmt.Errorf("provided queue manager namespace %v was not found on the currently logged-in cluster", flags.QueueManagerNamespace)
	}

	logger.Info("---- Starting Must-Gather tool ----")

	mustGatherToolStartTime := time.Now()

	// collect queue manager cr must-gathers
	logger.Info("---- Collecting queue manager details ----")
	fmt.Print("Collecting queue manager details ...")
	mustGatherStartTime := time.Now()
	err = gatherCRsToFiles(dynamicClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- queue manager details collected ----")

	// collect queue manager crd must-gathers
	logger.Info("---- Collecting queue manager CRD details ----")
	fmt.Print("Collecting queue manager CRD details ...")
	mustGatherStartTime = time.Now()
	err = gatherCRDsToFiles(dynamicClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- queue manager crd details collected ----")

	// collect pods must-gathers
	logger.Info("---- Collecting pod details ----")
	fmt.Print("Collecting pod details...")
	mustGatherStartTime = time.Now()
	err = gatherPodsToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- Pod details collected ----")

	// collect the route must-gathers
	logger.Info("---- Collecting route details ----")
	fmt.Print("Collecting route details...")
	mustGatherStartTime = time.Now()
	err = gatherRoutesToFiles(cfg, routeClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- Route details collected ----")

	// collect StatefulSet must-gathers
	logger.Info("---- Collecting StatefulSet details ----")
	fmt.Print("Collecting StatefulSet details...")
	mustGatherStartTime = time.Now()
	err = gatherStatefulSetToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- StatefulSet details collected ----")

	// collect Service must-gathers
	logger.Info("---- Collecting service details ----")
	fmt.Print("Collecting service details...")
	mustGatherStartTime = time.Now()
	err = gatherServicesToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("--- Service details collected ----")

	// collect PVC must-gathers
	logger.Info("---- Collecting PVC details ----")
	fmt.Print("Collecting PVC details...")
	mustGatherStartTime = time.Now()
	err = gatherPVCToFiles(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- PVC details collected ----")

	// collect MQ operator must-gathers
	logger.Info("---- Collecting mq-operator details ----")
	fmt.Print("Collecting mq-operator details...")
	mustGatherStartTime = time.Now()
	err = gatherMQOperatorToFiles(coreClient, dynamicClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- MQ-Operator details collected ----")

	// collect the cp4i csv details
	logger.Info("---- Collecting cp4i details ----")
	fmt.Print("Collecting cp4i details...")
	mustGatherStartTime = time.Now()
	err = gatherCp4IOperatorCSVToFiles(dynamicClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
	logger.Info("---- CP4I details collected ----")

	// if no-exec is disabled then collect the web-console and runmqras logs
	if !flags.NoExec {
		// collect web-console logs
		logger.Info("---- Collecting web-console details ----")
		fmt.Print("Collecting web-console details...")
		mustGatherStartTime = time.Now()
		err = gatherMQWebConsoleLogsToFiles(cfg, coreClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
		logger.Info("---- Web-console details collected ----")

		// collect runmqras logs
		logger.Info("---- Collecting runmqras details ----")
		fmt.Print("Collecting runmqras details(This may take time) ...")
		mustGatherStartTime = time.Now()
		err = gatherRunmqrasLogToFiles(cfg, coreClient, flags, logger)
		if err != nil {
			return err
		}
		fmt.Printf("Took: %v\n", time.Since(mustGatherStartTime))
		logger.Info("---- Runmqras details collected ----")
	}

	// delete the empty directories
	if err := utils.DeleteEmptyDirectories(flags.OutputDir, logger); err != nil {
		logger.Error(err.Error())
	}

	// if tar is enabled, then zip the must-gather output
	if flags.TarZip {
		fmt.Print("Compressing the logs...")
		mustGatherStartTime = time.Now()

		if err := tarzip.TarZipFolder(flags.OutputDir); err != nil {
			return fmt.Errorf("error tar zipping the collected must-gather at path %s: %v", flags.OutputDir, err)
		}

		fmt.Printf("Must Gathers archived. Took: %v\n", time.Since(mustGatherStartTime))
	}

	fmt.Printf("Must-Gather tool run completed... Took: %v\n", time.Since(mustGatherToolStartTime))
	logger.Info("---- Must-Gather tool run completed ----")

	fmt.Printf("Must-gather's collected, logs can be found at: %s\n", utils.GetLogFilePath(flags.OutputDir, utils.MustGatherLogFileName))

	return nil

}
