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
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/internal/mqagent"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
	"k8s.io/client-go/rest"
)

func MQAgent(args []string) error {

	// create flagset for "mqagent" os arg, and parse the args
	flags, err := parseFlags(args)
	if err != nil {
		if err.Error() == "help" {
			// Help was requested, so return after printing usage
			os.Exit(0)
		}
		return fmt.Errorf("Error parsing mq-agent mustgather flags: %v", err)
	}

	// handle non-required flag defaults
	if err := setDefaultFlags(&flags); err != nil {
		return err
	}

	// generate the kubernetes config
	cfg, err := kubeclient.BuildKubeConfig(flags.KubeconfigPath)
	if err != nil {
		// If kubeconfigPath doesn't work attempt to use inClusterConfig
		// If InClusterConfig succeed mq-container-inspector is probably being run from a pod
		icConfig, icError := rest.InClusterConfig()
		if icError == nil {
			cfg = icConfig
		} else {
			// if there is an error with InClusterConfig(),
			//  we want to return the BuildKubeConfig error.
			return fmt.Errorf("error building kube-config: %v", err)
		}
	}

	if err := mqagent.MQAgentMustGather(cfg, flags); err != nil {
		return err
	}

	return nil
}

func parseFlags(args []string) (utils.MQAgentFlags, error) {

	var flags utils.MQAgentFlags

	flagSet := flag.NewFlagSet(utils.MQAgents, flag.ContinueOnError)

	flagSet.StringVar(&flags.AgentReleaseName, "agent-release-name", "", "name of the mq-agent release(required)")
	flagSet.StringVar(&flags.Namespace, "namespace", "", "namespace where the mq-agent is deployed(required)")
	flagSet.StringVar(&flags.KubeconfigPath, "kubeconfig", "", "path to the kubeconfig file. Ignored when running via mustgather image (default: ~/.kube/config)")
	flagSet.StringVar(&flags.OutputDir, "output-dir", "", "directory where the must-gather output folder will be created. Ignored when running via mustgather image (default: current working directory)")
	flagSet.BoolVar(&flags.SkipTar, "skip-tar", false, "skip compressing the mq-agent must-gather output into a tar.gz file (default: false)")
	flagSet.BoolVar(&flags.SkipExec, "skip-exec", false, "skip gathering data that require container exec access, e.g. dmp, trc files (default: false)")
	flagSet.BoolVar(&flags.Help, "help", false, "show help message")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", flagSet.Name())
		flagSet.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "--%s\n        %s\n", f.Name, f.Usage)
		})
	}

	err := flagSet.Parse(args)
	if err != nil {
		return flags, err
	}

	if flags.Help {
		flagSet.Usage()
		return flags, fmt.Errorf("help")
	}

	if len(flag.Args()) > 0 {
		return flags, fmt.Errorf("unexpected arguments: %v", flagSet.Args())
	}

	// validate if required flags have been passed
	if ok, message := validateRequiredFlags(flags); !ok {
		flagSet.Usage()
		return flags, fmt.Errorf("%s", message)
	}

	return flags, nil

}

func setDefaultFlags(flags *utils.MQAgentFlags) error {

	// setup the KubeconfigPath
	if flags.KubeconfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error retrieving user's home directory: %v", err)
		}
		flags.KubeconfigPath = filepath.Join(homeDir, ".kube/config")
	}

	// setup the outputDir
	if flags.OutputDir == "" {
		currentWorkingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting user's current working directory: %v", err)
		}
		flags.OutputDir = currentWorkingDir
	}

	// create a directory inside the output directory, because we don't want to zip the entire outputDir, as it may contain other files apart from must-gather details
	timestamp := utils.GetCurrentTimestamp(utils.TimestampFormat)
	outputDir := filepath.Join(flags.OutputDir, fmt.Sprintf("MQ-Agent_MustGather_%v", timestamp))
	flags.OutputDir = outputDir

	return nil

}

func validateRequiredFlags(flags utils.MQAgentFlags) (bool, string) {

	if flags.AgentReleaseName == "" {
		return false, "--agent-release-name is required"
	}

	if flags.Namespace == "" {
		return false, "--namespace is required"
	}

	return true, ""

}
