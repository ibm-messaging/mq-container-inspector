package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	controllerrevisions "github.ibm.com/mq-cloudpak/mq-inspector/pkg/controller_revisions"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/daemonset"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func gatherDaemonSetToFiles(client kubernetes.Interface, flags utils.MustGatherFlags, qmPod *corev1.Pod, logger *slog.Logger) error {

	daemonSetDirectory := filepath.Join(flags.OutputDir, "daemonsets")
	if !utils.CheckIfDirectoryExist(daemonSetDirectory) {
		if err := utils.CreateDirectory(daemonSetDirectory, 0775); err != nil {
			return err
		}
	}

	daemonSetList, daemonSetControllerRevisionList, daemonSetEvents := getDaemonSetDetailsByPodName(client, flags, qmPod, logger)

	// write the daemonset details to it's own yaml file
	daemonSetDetailsFileNameFormat := "%s-daemonset.yaml"
	if err := daemonset.WriteDaemonSetsToFile(daemonSetList, daemonSetDetailsFileNameFormat, daemonSetDirectory); err != nil {
		logger.Error(err.Error())
	}

	// write the daemonset revisions to their yamls
	daemonSetRevisionsFileNameFormat := "%s-daemonset-revisions.yaml"
	if err := controllerrevisions.WriteControllerRevisionYamlsToFile(daemonSetControllerRevisionList, daemonSetRevisionsFileNameFormat, daemonSetDirectory); err != nil {
		logger.Error(err.Error())
	}

	// write the daemonset events to their files
	daemonSetEventsFileNameFormat := "%s-daemonset-events.txt"
	if err := daemonset.WriteDaemonSetEventsToFile(daemonSetEvents, daemonSetEventsFileNameFormat, daemonSetDirectory); err != nil {
		logger.Error(err.Error())
	}

	if fileCount, err := utils.GetFileCountInDirectory(daemonSetDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("DaemonSet details: %s: Total Files: %d", daemonSetDirectory, fileCount))
	}

	return nil

}

func getDaemonSetDetailsByPodName(client kubernetes.Interface, flags utils.MustGatherFlags, qmPod *corev1.Pod, logger *slog.Logger) ([]appsv1.DaemonSet, []appsv1.ControllerRevision, map[string][]corev1.Event) {

	// get the daemonset details by pod name
	daemonSetDetails, err := daemonset.GetDaemonSetByPodName(client, qmPod.ObjectMeta.Name, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error fetching daemon set details for %s pod: %v", qmPod.ObjectMeta.Name, err))
	}
	daemonSetList := []appsv1.DaemonSet{*daemonSetDetails}

	// get the controller revision details by daemonset selector
	daemonSetLabels := daemonSetDetails.ObjectMeta.Labels
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: daemonSetLabels})

	daemonSetControllerRevisionList, err := controllerrevisions.GetControllerRevisionsBySelector(client, labelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Warn(fmt.Sprintf("error while fetching daemonset revisions with selector %s: %v", labelSelector, err))
	}

	// get the daemonset events by daemonset name
	daemonSetEvents, err := daemonset.GetDaemonSetEventsByName(client, daemonSetDetails.ObjectMeta.Name, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching daemonset events with name %s: %v", daemonSetDetails.ObjectMeta.Name, err))
	}

	return daemonSetList, daemonSetControllerRevisionList, daemonSetEvents
}
