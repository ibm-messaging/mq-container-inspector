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
package utils

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
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
	NoExec                bool   //whether to collect the pod-exec must-gathers like runmqras etc.
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

var CheckIfDirectoryExist = func(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

var CreateDirectory = func(dir string, perm fs.FileMode) error {
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("error creating output-directory(%s): %v", dir, err)
	}
	return nil
}

var GetLogFilePath = func(outputDirectory, logFileName string) string {
	return filepath.Join(outputDirectory, logFileName)
}

var InitializeLogFile = func(outputDirectory, logFileName string) (*os.File, error) {
	logFilePath := GetLogFilePath(outputDirectory, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return nil, fmt.Errorf("error creating logfile %s in base-directory %s: %v", logFileName, outputDirectory, err)
	}
	return logFile, nil

}

var GetFileCountInDirectory = func(directoryName string) (int, error) {

	entries, err := os.ReadDir(directoryName)
	if err != nil {
		return 0, fmt.Errorf("error while reading %s directory: %v", directoryName, err)
	}

	fileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			fileCount++
		}
	}

	return fileCount, nil

}

var DeleteEmptyDirectories = func(directoryName string, logger *slog.Logger) error {

	entries, err := os.ReadDir(directoryName)
	if err != nil {
		return fmt.Errorf("error while reading %s directory", directoryName)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			entryPath := filepath.Join(directoryName, entry.Name())
			fileCount, err := GetFileCountInDirectory(entryPath)
			if err != nil {
				return err
			}

			if fileCount == 0 {
				if err := os.RemoveAll(entryPath); err != nil {
					return err
				} else {
					logger.Info(fmt.Sprintf("Deleting empty directory: %s/%s", directoryName, entry.Name()))
				}
			}
		}
	}

	return nil

}
