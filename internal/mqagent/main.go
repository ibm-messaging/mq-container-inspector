/*
© Copyright IBM Corporation 2026

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

package mqagent

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/kubeclient"
	mqagent "github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/mq-agent"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/tarzip"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/validations"
	"k8s.io/client-go/rest"
)

func MQAgentMustGather(cfg *rest.Config, flags utils.MQAgentFlags) error {

	// build the required clients
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	routeClient, err := kubeclient.BuildRouteClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building route client from config: %v", err)
	}

	// check if the provided namespace exists on the cluster
	if err := validations.ValidateNamespace(coreClient, flags.Namespace); err != nil {
		return err
	}

	// If OutputDir doesnot exist create the directory
	directoryExists := utils.CheckIfDirectoryExist(flags.OutputDir)
	if !directoryExists {
		if err := utils.CreateDirectory(flags.OutputDir, 0o775); err != nil {
			return err
		}
	}

	// initialize logger
	logFile, err := utils.InitializeLogFile(flags.OutputDir, utils.MQAgentMustGatherLogFileName)
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

	logger.Info("---- Starting MQ-Agent must-gather tool ----")

	mqAgentMustGatherToolStartTime := time.Now()

	logger.Info("---- Collecting Deployment details ----")
	fmt.Print("Collecting Deployment details...")
	mustGatherTime := time.Now()
	deploymentNameList, err := mqagent.CollectMQAgentDeploymentDetails(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherTime).Round(time.Millisecond))
	logger.Info("---- Deployment details collected ----")

	logger.Info("---- Collecting ReplicaSet details ----")
	fmt.Print("Collecting ReplicaSet details...")
	mustGatherTime = time.Now()
	replicaSetNameList, err := mqagent.CollectMQAgentReplicaSetDetails(coreClient, flags, deploymentNameList, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherTime).Round(time.Millisecond))
	logger.Info("---- ReplicaSet details collected ----")

	logger.Info("---- Collecting Pod details(This may take some time) ----")
	fmt.Print("Collecting Pod details(This may take some time)...")
	mustGatherTime = time.Now()
	err = mqagent.CollectMQAgentPodDetails(coreClient, flags, replicaSetNameList, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherTime).Round(time.Millisecond))
	logger.Info("---- Pod details collected ----")

	logger.Info("---- Collecting Service details ----")
	fmt.Print("Collecting Service details...")
	mustGatherTime = time.Now()
	serviceNameList, err := mqagent.CollectMQAgentServiceDetails(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherTime).Round(time.Millisecond))
	logger.Info("---- Service details collected ----")

	logger.Info("---- Collecting Route details ----")
	fmt.Print("Collecting Route details...")
	mustGatherTime = time.Now()
	err = mqagent.CollectMQAgentRouteDetails(routeClient, serviceNameList, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherTime).Round(time.Millisecond))
	logger.Info("---- Route details collected ----")

	logger.Info("---- Collecting Network Policies ----")
	fmt.Print("Collecting Network Policies details...")
	mustGatherTime = time.Now()
	err = mqagent.CollectMQAgentNetworkPolicies(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherTime).Round(time.Millisecond))
	logger.Info("---- Network Policies details collected ----")

	logger.Info("---- Collecting ServiceAccount details ----")
	fmt.Print("Collecting ServiceAccount details ...")
	mustGatherTime = time.Now()
	err = mqagent.CollectMQAgentServiceAccountDetails(coreClient, flags, logger)
	if err != nil {
		return err
	}
	fmt.Printf("Took: %v\n", time.Since(mustGatherTime).Round(time.Millisecond))
	logger.Info("---- ServiceAccount details collected ----")

	// delete the empty directories
	if err := utils.DeleteEmptyDirectories(flags.OutputDir, logger); err != nil {
		logger.Error(err.Error())
	}

	// if skip-tar is disabled, then tar the mq-agent must-gather output
	if !flags.SkipTar {
		fmt.Print("Compressing the logs...")
		mustGatherStartTime := time.Now()

		if err := tarzip.TarZipFolder(flags.OutputDir); err != nil {
			return fmt.Errorf("error tar zipping the collected mq-agent must-gather at path %s: %v", flags.OutputDir, err)
		}

		fmt.Printf("MQ-Agent Must Gathers archived. Took: %v\n", time.Since(mustGatherStartTime).Round(time.Millisecond))
	}

	fmt.Printf("MQ-Agent must-gather tool run completed... Took: %v\n", time.Since(mqAgentMustGatherToolStartTime).Round(time.Millisecond))
	logger.Info("---- MQ-Agent must-gather tool run completed ----")

	return nil

}
