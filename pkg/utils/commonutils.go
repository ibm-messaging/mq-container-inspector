/*
© Copyright IBM Corporation 2025, 2026

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

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
)

type MustGatherFlags struct {
	QueueManagerName      string // QueueManager resource name
	PodName               string // pod name
	QueueManagerNamespace string // QueueManager  resource namespace
	OperatorNamespace     string // namespace where the MQ Operator is running
	QmContainerName       string // QueueManager container name
	KubeconfigPath        string // path to the kubeconfig file
	OutputDir             string // directory where the must-gathers will be stored
	SkipTar               bool   // whether to tar+zip the must-gather output
	SkipExec              bool   // whether to collect the pod-exec must-gathers like runmqras etc.
	Help                  bool   // Flag to display help message
}

type PVCInspectorFlags struct {
	QueueManagerName      string // QueueManager resource name
	PodName               string // pod name
	QueueManagerNamespace string // QueueManager  resource namespace
	KubeconfigPath        string // path to the kubeconfig file
	OutputDir             string // directory where the collected pvc-inspector data will be stored
	Cleanup               bool   // whether to cleanup the pvc-pods at the end of the tool run
	DryRun                bool   // whether to create the pvc-pods
	SkipTar               bool   // whether to tar+zip the pvc-inspector output
	Help                  bool   // Flag to display help message
	Runmqras              bool   // whether to run the runmqras command automatically
}

type MQAgentFlags struct {
	AgentReleaseName string // Release name for MQ-Agent
	Namespace        string // namespace where agent is installed
	KubeconfigPath   string // path to the kubeconfig file
	OutputDir        string // directory where the collected mq-agent must-gather data will be stored
	SkipTar          bool   // whether to tar+zip the pvc-inspector output
	Help             bool   // Flag to display help message
}

func GetCurrentTimestamp(timeFormat string) string {
	return time.Now().Format(timeFormat)
}

type PodLogs struct {
	PreviousPodLogsRequest *rest.Request
	CurrentPodLogsRequest  *rest.Request
}

func FetchQMGRResourceNameFromSelector(selector string) (string, error) {
	if index := strings.Index(selector, "="); index != -1 {
		return selector[index+1:], nil
	}

	return "", fmt.Errorf("error invalid selector format: %s", selector)
}

func FormatFilePath(outputDir, fileNameFormat string, args ...interface{}) string {
	filename := fmt.Sprintf(fileNameFormat, args...)
	return filepath.Join(outputDir, filename)
}

type ExecConfig struct {
	KubernetesConfig *rest.Config
	PodName          string
	Namespace        string
	ContainerName    string
	Cmd              []string
	Logger           *slog.Logger
}

func CheckIfDirectoryExist(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func CreateDirectory(dir string, perm fs.FileMode) error {
	if err := SafeMkdirAll(dir, ".", perm); err != nil {
		return fmt.Errorf("error creating output-directory(%s): %v", dir, err)
	}
	return nil
}

func GetLogFilePath(outputDirectory, logFileName string) string {
	return filepath.Join(outputDirectory, logFileName)
}

func InitializeLogFile(outputDirectory, logFileName string) (*os.File, error) {

	logFile, err := SafeOpenFile(outputDirectory, logFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("error creating logfile %s in base-directory %s: %v", logFileName, outputDirectory, err)
	}
	return logFile, nil

}

func GetFileCountInDirectory(directoryName string) (int, error) {

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

func DeleteEmptyDirectories(directoryName string, logger *slog.Logger) error {

	baseDir := filepath.Clean(directoryName)

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
				if err := SafeRemoveAll(baseDir, entry.Name()); err != nil {
					return err
				} else {
					logger.Info(fmt.Sprintf("Deleting empty directory: %s/%s", directoryName, entry.Name()))
				}
			}
		}
	}

	return nil

}

func GetPodInstance(pod *corev1.Pod) string {

	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == NativeHAEnvName && env.Value == "true" {
				return NativeHA
			} else if env.Name == MultiInstanceEnvName && env.Value == "true" {
				return MultiInstance
			}
		}
	}
	return SingleInstance
}

func IsSelectorSubsetOfLabels(selector, label map[string]string) bool {

	for key, value := range selector {
		if label[key] != value {
			return false
		}
	}

	return true

}

func GetPodOwner(pod *corev1.Pod) string {

	// a pod can have multiple owners but only a single controlling owner which is what we return
	for _, owner := range pod.ObjectMeta.OwnerReferences {
		if owner.Kind != "" && *owner.Controller {
			return owner.Kind
		}
	}

	return ""
}

func GetPVCInspectorContainerCommand() []string {
	return []string{"tail", "-f", "/dev/null"}
}

func CheckIfPodHasPersistedStorage(pod corev1.Pod) bool {

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			return true
		}
	}

	return false

}

func IsQueueManagerPod(pod *corev1.Pod) bool {

	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == QueueManagerEnvName && env.Value != "" {
				return true
			}
		}
	}

	return false

}

func RemoveFile(file *os.File) error {

	if err := os.Remove(file.Name()); err != nil {
		return err
	}

	return nil

}

func SafeJoinUnderBase(baseDir, userPath string) (string, error) {

	if baseDir == "" {
		return "", fmt.Errorf("base directory must not be empty")
	}

	// Normalize the baseDir to absolute and cleanPath
	baseAbs, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute base dir %q: %w", baseDir, err)
	}

	// Helper to cjheck if candidate is under base
	isCandidateUnderBase := func(baseAbs, candidateAbs string) bool {
		sep := string(os.PathSeparator)
		prefix := baseAbs
		if !strings.HasSuffix(prefix, sep) {
			prefix += sep
		}

		return candidateAbs == baseAbs || strings.HasPrefix(candidateAbs, prefix)
	}

	// If userPath is absolute (or resolves absolute) and already under base, accept it
	candidateAbs, err := filepath.Abs(filepath.Clean(userPath))
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %q: %w", userPath, err)
	}

	if isCandidateUnderBase(baseAbs, candidateAbs) {
		return candidateAbs, nil
	}

	// Join and normalize the candidate path
	joined := filepath.Join(baseAbs, userPath)
	// Otherwise treat userPath as relative-to-base
	candidateAbs, err = filepath.Abs(filepath.Clean(joined))
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %q: %w", candidateAbs, err)
	}

	// Enforce candidate is under baseAbs
	if !isCandidateUnderBase(baseAbs, candidateAbs) {
		return "", fmt.Errorf("path %q escapes base directory %q", candidateAbs, baseAbs)
	}

	return candidateAbs, nil

}

func SafeMkdirAll(baseDir, rel string, perm os.FileMode) error {

	safePath, err := SafeJoinUnderBase(baseDir, rel)
	if err != nil {
		return err
	}

	return os.MkdirAll(safePath, perm)

}

func SafeOpenFile(baseDir, rel string, flags int, perm os.FileMode) (*os.File, error) {

	safePath, err := SafeJoinUnderBase(baseDir, rel)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(safePath, flags, perm)

}

func SafeRemoveAll(baseDir, relativePath string) error {

	safePath, err := SafeJoinUnderBase(baseDir, relativePath)
	if err != nil {
		return err
	}

	baseAbs, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return err
	}

	if safePath == baseAbs {
		return fmt.Errorf("cannot remove base directory %q", baseAbs)
	}

	return os.RemoveAll(safePath)

}

func FetchPodContainers(pods []corev1.Pod) map[string][]string {

	podContainerNameMap := make(map[string][]string)

	for _, pod := range pods {
		var containerNames []string
		for _, container := range pod.Spec.Containers {
			containerNames = append(containerNames, container.Name)
		}
		podContainerNameMap[pod.ObjectMeta.Name] = containerNames
	}

	return podContainerNameMap

}

func FilterDeploymentListByAnnotation(deploymentList []appsv1.Deployment, annotationKey string, annotationValue string) []appsv1.Deployment {
	var deployments []appsv1.Deployment

	for _, deployment := range deploymentList {

		annotations := deployment.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			deployments = append(deployments, deployment)
		}

	}

	return deployments
}

func FilterReplicasetListByAnnotation(replicasetList []appsv1.ReplicaSet, annotationKey string, annotationValue string) []appsv1.ReplicaSet {

	var replicasets []appsv1.ReplicaSet

	for _, replicaset := range replicasetList {

		annotations := replicaset.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			replicasets = append(replicasets, replicaset)
		}
	}

	return replicasets
}

func FilterServiceListByAnnotation(servicesList []corev1.Service, annotationKey string, annotationValue string) []corev1.Service {

	var services []corev1.Service

	for _, service := range servicesList {
		annotations := service.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			services = append(services, service)
		}
	}

	return services
}

func FilterRouteListByAnnotation(routeList []routev1.Route, annotationKey string, annotationValue string) []routev1.Route {

	var routes []routev1.Route

	for _, route := range routeList {
		annotations := route.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			routes = append(routes, route)
		}
	}

	return routes
}

func FilterNetworkPolicyListByAnnotation(networkPolicyList []networkv1.NetworkPolicy, annotationKey, annotationValue string) []networkv1.NetworkPolicy {

	var filteredNetworkPolicyList []networkv1.NetworkPolicy

	for _, networkPolicy := range networkPolicyList {

		annotations := networkPolicy.Annotations

		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			filteredNetworkPolicyList = append(filteredNetworkPolicyList, networkPolicy)
		}
	}

	return filteredNetworkPolicyList

}

func FilterServiceAccountListByAnnotation(serviceAccountList []corev1.ServiceAccount, annotationKey string, annotationValue string) []corev1.ServiceAccount {

	var serviceAccounts []corev1.ServiceAccount

	for _, serviceAccount := range serviceAccountList {
		annotations := serviceAccount.Annotations

		// check if release-name annotation matches
		if value, ok := annotations[annotationKey]; ok && value == annotationValue {
			serviceAccounts = append(serviceAccounts, serviceAccount)
		}
	}

	return serviceAccounts
}
