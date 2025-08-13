package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	controllerrevisions "github.ibm.com/mq-cloudpak/mq-inspector/pkg/controller_revisions"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/deployment"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func gatherDeploymentToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags, qmPod *corev1.Pod, logger *slog.Logger) error {

	// create the deployment directory
	deploymentDirectory := filepath.Join(flags.OutputDir, "deployments")
	if !utils.CheckIfDirectoryExist(deploymentDirectory) {
		if err := utils.CreateDirectory(deploymentDirectory, 0775); err != nil {
			return err
		}
	}

	deploymentList, deploymentControllerRevisionList, deploymentEvents := getDeploymentDetailsByPodName(coreClient, flags, qmPod, logger)

	// write the deployment details to it's own yaml file
	fileNameFormat := "%s-deployment.yaml"
	if err := deployment.WriteDeploymentsToFile(deploymentList, fileNameFormat, deploymentDirectory); err != nil {
		logger.Error(err.Error())
	}

	// write the deployment revisions to their yamls
	deploymentRevisionsFileNameFormat := "%s-deployment-revisions.yaml"
	if err := controllerrevisions.WriteControllerRevisionYamlsToFile(deploymentControllerRevisionList, deploymentRevisionsFileNameFormat, deploymentDirectory); err != nil {
		logger.Error(err.Error())
	}

	// write the deployment events to their files
	deploymentEventsFileNameFormat := "%s-deployment-events.txt"
	if err := deployment.WriteDeploymentEventsToFile(deploymentEvents, deploymentEventsFileNameFormat, deploymentDirectory); err != nil {
		logger.Error(err.Error())
	}

	if fileCount, err := utils.GetFileCountInDirectory(deploymentDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("Deployment details: %s: Total Files: %d", deploymentDirectory, fileCount))
	}

	return nil

}

func getDeploymentDetailsByPodName(coreClient kubernetes.Interface, flags utils.MustGatherFlags, qmPod *corev1.Pod, logger *slog.Logger) ([]appsv1.Deployment, []appsv1.ControllerRevision, map[string][]corev1.Event) {

	// get the deployment details by the deployment name
	deploymentDetails, err := deployment.GetDeploymentsByPodName(coreClient, qmPod.ObjectMeta.Name, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error fetching deployment details for %s pod: %v", qmPod.ObjectMeta.Name, err))
	}

	deploymentList := []appsv1.Deployment{*deploymentDetails}

	// get the deployment controller revision list
	deploymentLabels := deploymentDetails.ObjectMeta.Labels
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: deploymentLabels})

	deploymentControllerRevisionList, err := controllerrevisions.GetControllerRevisionsBySelector(coreClient, labelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching deployment revisions with selector %s: %v", labelSelector, err))
	}

	// get the deployment events by deployment name
	deploymentEvents, err := deployment.GetDeploymentEventsByName(coreClient, deploymentDetails.ObjectMeta.Name, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching deployment events with name %s: %v", deploymentDetails.ObjectMeta.Name, err))
	}

	return deploymentList, deploymentControllerRevisionList, deploymentEvents

}
