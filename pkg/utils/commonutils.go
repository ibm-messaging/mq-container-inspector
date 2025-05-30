package utils

import (
	"time"
)

type MustGatherFlags struct {
	QueueManagerName      string // QueueManager resource name
	QueueManagerNamespace string // QueueManager  resource namespace
	OperatorNamespace     string // namespace where the MQ Operator is running
	KubeconfigPath        string // path to the kubeconfig file
	OutputDir             string // directory where the must-gathers will be stored
	TarZip                bool   // whether to tar+zip the must-gather output
}

var GetCurrentTimestamp = func(timeFormat string) string {
	return time.Now().Format(timeFormat)
}
