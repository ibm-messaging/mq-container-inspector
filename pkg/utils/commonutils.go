package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/rest"
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

type PodLogs struct {
	PreviousPodLogsRequest *rest.Request
	CurrentPodLogsRequest  *rest.Request
}

var FetchQMGRResourceNameFromSelector = func(selector string) (string, error) {
	if index := strings.Index(selector, "="); index != -1 {
		return selector[index+1:], nil
	}

	return "", fmt.Errorf("error invalid selector format: %s", selector)
}

var FormatFilePath = func(outputDir, fileNameFormat string, args ...interface{}) string {
	filename := fmt.Sprintf(fileNameFormat, args...)
	return filepath.Join(outputDir, filename)
}

type ExecConfig struct {
	KubernetesConfig *rest.Config
	PodName          string
	Namespace        string
	ContainerName    string
	Cmd              []string
}
