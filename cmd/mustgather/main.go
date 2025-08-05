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
package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-inspector/internal/mustgather"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
)

func MustGather(args []string) error {

	// create flagset for "mustgather" os arg, and parse the args
	flags, err := parseFlags(args)
	if err != nil {
		return fmt.Errorf("error parsing mustgather flags: %v", err)
	}

	// handle non-required flag defaults
	if err := setDefaultFlags(&flags); err != nil {
		return err
	}

	// generate the kubernetes config
	cfg, err := kubeclient.BuildKubeConfig(flags.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("error building kube-config: %v", err)
	}

	if err := mustgather.MustGather(cfg, flags); err != nil {
		return err
	}

	return nil

}

func parseFlags(args []string) (utils.MustGatherFlags, error) {

	var flags utils.MustGatherFlags

	flagSet := flag.NewFlagSet(utils.MustGather, flag.ContinueOnError)

	flagSet.StringVar(&flags.QueueManagerName, "qm-name", "", "QueueManager custom resource metadata.name")
	flagSet.StringVar(&flags.PodName, "pod-name", "", "pod name of the mq instance")
	flagSet.StringVar(&flags.QueueManagerNamespace, "qm-namespace", "", "QueueManager custom resource metadata.namespace (required)")
	flagSet.StringVar(&flags.OperatorNamespace, "operator-namespace", "", "MQ Operator namespace")
	flagSet.StringVar(&flags.KubeconfigPath, "kubeconfig", "", "kubeconfig file path, defaults to '.kube/config'")
	flagSet.StringVar(&flags.OutputDir, "output-dir", "", "output directory where the must-gather files will be stored, defaults to current-working-directory")
	flagSet.BoolVar(&flags.TarZip, "tar-zip", true, "whether or not to tar-zip the must-gathers, defaults to true")
	flagSet.BoolVar(&flags.NoExec, "no-exec", false, "disable must-gathers commands that require container exec access (e.g. runmqras, webconsole logs), defaults to false")
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flagSet.PrintDefaults()
	}

	err := flagSet.Parse(args)
	if err != nil {
		return flags, err
	}

	// validate if required flags have been passed
	if ok, message := validateRequiredFlags(flags); !ok {
		flagSet.Usage()
		return flags, fmt.Errorf("%s", message)
	}

	return flags, nil

}

func setDefaultFlags(flags *utils.MustGatherFlags) error {

	if flags.KubeconfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error retrieving user's home directory: %v", err)
		}
		flags.KubeconfigPath = filepath.Join(homeDir, ".kube/config")
	}

	timestamp := utils.GetCurrentTimestamp(utils.TimestampFormat)
	if flags.OutputDir == "" {
		currentWorkingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting user's current working directory: %v", err)
		}

		flags.OutputDir = filepath.Join(currentWorkingDir, fmt.Sprintf("Must_Gather_%v", timestamp))
		// If OutputDir doesnot exist create the directory
		directoryExist := utils.CheckIfDirectoryExist(flags.OutputDir)
		if !directoryExist {
			if err := utils.CreateDirectory(flags.OutputDir, 0775); err != nil {
				return err
			}
		}
		return nil
	}

	// create a directory inside the output directory, because we don't want to zip the entire outputDir, as it may contain other files apart from must-gather details
	outputDir := filepath.Join(flags.OutputDir, fmt.Sprintf("Must_Gather_%v", timestamp))
	flags.OutputDir = outputDir

	// If OutputDir doesnot exist create the directory
	directoryExist := utils.CheckIfDirectoryExist(flags.OutputDir)
	if !directoryExist {
		if err := utils.CreateDirectory(flags.OutputDir, 0775); err != nil {
			return err
		}
	}

	return nil

}

func validateRequiredFlags(flags utils.MustGatherFlags) (bool, string) {

	if flags.QueueManagerNamespace == "" {
		return false, "error: --qm-namespace is required"
	}

	if flags.QueueManagerName == "" && flags.PodName == "" {
		return false, "error: at least one of --qm-name or --pod-name must be provided"
	}

	if flags.QueueManagerName != "" && flags.PodName != "" {
		return false, "error: only one of --qm-name or --pod-name should be provided"
	}

	return true, ""
}
