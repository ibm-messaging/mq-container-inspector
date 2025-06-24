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
		return fmt.Errorf("error parsing mustgather flags: %v\n", err)
	}

	// handle non-required flag defaults
	if err := setDefaultFlags(&flags); err != nil {
		return err
	}

	// generate the kubernetes config
	cfg, err := kubeclient.BuildKubeConfig(flags.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("error building kube-config: %v\n", err)
	}

	if err := mustgather.MustGather(cfg, flags); err != nil {
		return fmt.Errorf("error running must gather: %v\n", err)
	}

	return nil

}

func parseFlags(args []string) (utils.MustGatherFlags, error) {

	var flags utils.MustGatherFlags

	flagSet := flag.NewFlagSet(utils.MustGather, flag.ContinueOnError)

	flagSet.StringVar(&flags.QueueManagerName, "qm-name", "", "QueueManager custom resource metadata.name (required)")
	flagSet.StringVar(&flags.QueueManagerNamespace, "qm-namespace", "", "QueueManager custom resource metadata.namespace (required)")
	flagSet.StringVar(&flags.OperatorNamespace, "operator-namespace", "", "MQ Operator namespace")
	flagSet.StringVar(&flags.KubeconfigPath, "kubeconfig", "", "kubeconfig file path, defaults to '.kube/config'")
	flagSet.StringVar(&flags.OutputDir, "output-dir", "", "output directory where the must-gather files will be stored, defaults to current-working-directory")
	flagSet.BoolVar(&flags.TarZip, "tar-zip", true, "whether or not to tar-zip the must-gathers, defaults to true")
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flagSet.PrintDefaults()
	}

	err := flagSet.Parse(args)
	if err != nil {
		return flags, err
	}

	// validate if required flags have been passed
	if !validateRequiredFlags(flags) {
		flagSet.Usage()
		return flags, fmt.Errorf("error required flags are missing")
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

	if flags.OutputDir == "" {
		currentWorkingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting user's current working directory: %v", err)
		}

		timestamp := utils.GetCurrentTimestamp(utils.TimestampFormat)
		flags.OutputDir = filepath.Join(currentWorkingDir, fmt.Sprintf("Must_Gather_%v", timestamp))
	}

	// If OutputDir doesnot exist create the directory
	if _, err := os.Stat(flags.OutputDir); os.IsNotExist(err) {
		err = os.Mkdir(flags.OutputDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating output-directory(%s): %v", flags.OutputDir, err)
		}
	}

	return nil

}

func validateRequiredFlags(flags utils.MustGatherFlags) bool {

	if flags.QueueManagerName == "" || flags.QueueManagerNamespace == "" {
		return false
	}

	return true
}
