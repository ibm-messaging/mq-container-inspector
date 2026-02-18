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
package mustgather

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/statefulset"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"

	"k8s.io/client-go/kubernetes"

	controllerrevisions "github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/controller_revisions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func gatherStatefulSetToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) error {

	// create StatefulSet directory to store StatefulSet files
	statefulSetDirectory := filepath.Join(flags.OutputDir, "statefulsets")
	directoryExist := utils.CheckIfDirectoryExist(statefulSetDirectory)
	if !directoryExist {
		if err := utils.CreateDirectory(statefulSetDirectory, 0775); err != nil {
			return err
		}
	}

	var statefulSetList []appsv1.StatefulSet
	var statefulSetRevisionList []appsv1.ControllerRevision
	var statefulSetEvents map[string][]corev1.Event
	var err error

	if flags.QueueManagerName != "" {

		statefulSetList, statefulSetRevisionList, statefulSetEvents, err = getStatefulSetDetailsBySelector(coreClient, flags, logger)
		if err != nil {
			return err
		}

	} else if flags.PodName != "" {

		statefulSetList, statefulSetRevisionList, statefulSetEvents, err = getStatefulSetDetailsByPodName(coreClient, flags, logger)
		if err != nil {
			return err
		}

	}

	// write the StatefulSet details to their yamls
	statefulSetDetailsFileNameFormat := "%s-statefulset.yaml"
	if err := statefulset.WriteStatefulSetYamlsToFile(statefulSetList, statefulSetDetailsFileNameFormat, statefulSetDirectory); err != nil {
		logger.Error(err.Error())
	}

	// write the StatefulSet revisions to their yamls
	statefulSetRevisionsFileNameFormat := "%s-statefulset-revisions.yaml"
	if err := controllerrevisions.WriteControllerRevisionYamlsToFile(statefulSetRevisionList, statefulSetRevisionsFileNameFormat, statefulSetDirectory); err != nil {
		logger.Error(err.Error())
	}

	// write the StatefulSet events to their files
	statefulSetEventsFileNameFormat := "%s-statefulset-events.txt"
	if err := statefulset.WriteStatefulSetEventsToFile(statefulSetEvents, statefulSetEventsFileNameFormat, statefulSetDirectory); err != nil {
		logger.Error(err.Error())
	}

	if fileCount, err := utils.GetFileCountInDirectory(statefulSetDirectory); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info(fmt.Sprintf("StatefulSet details: %s: Total Files: %d", statefulSetDirectory, fileCount))
	}

	return nil

}

func getStatefulSetDetailsBySelector(coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) ([]appsv1.StatefulSet, []appsv1.ControllerRevision, map[string][]corev1.Event, error) {
	statefulSetLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get StatefulSet details by selector
	statefulSetList, err := statefulset.GetStatefulSetDetailsBySelector(coreClient, statefulSetLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching StatefulSets with selector %s: %v", statefulSetLabelSelector, err))
	}

	// get StatefulSet revisions by selector
	statefulSetRevisionList, err := controllerrevisions.GetControllerRevisionsBySelector(coreClient, statefulSetLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Warn(fmt.Sprintf("error while fetching StatefulSet revisions with selector %s: %v", statefulSetLabelSelector, err))
	}

	// get StatefulSet events by selector
	statefulSetEvents, err := statefulset.GetStatefulSetEventsBySelector(coreClient, statefulSetLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching StatefulSet events with selector %s: %v", statefulSetLabelSelector, err))
	}

	return statefulSetList, statefulSetRevisionList, statefulSetEvents, nil
}

func getStatefulSetDetailsByPodName(coreClient kubernetes.Interface, flags utils.MustGatherFlags, logger *slog.Logger) ([]appsv1.StatefulSet, []appsv1.ControllerRevision, map[string][]corev1.Event, error) {
	// get StatefulSet details by pod name
	statefulSetDetails, err := statefulset.GetStatefulSetDetailsByPodName(coreClient, flags.PodName, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching StatefulSets for %s pod: %v", flags.PodName, err))
	}
	statefulSetList := []appsv1.StatefulSet{*statefulSetDetails}

	// get the StatefulSet revisions by pod name
	// we can filter them by the retrieved StatefulSet labels
	statefulSetLabels := statefulSetDetails.ObjectMeta.Labels
	statefulSetLabelString := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: statefulSetLabels})

	statefulSetRevisionList, err := controllerrevisions.GetControllerRevisionsBySelector(coreClient, statefulSetLabelString, flags.QueueManagerNamespace)
	if err != nil {
		logger.Warn(fmt.Sprintf("error while fetching StatefulSet revisions with selector %s: %v", statefulSetLabelString, err))
	}

	// get StatefulSet events by StatefulSet name
	statefulSetEvents, err := statefulset.GetStatefulSetEventsByName(coreClient, statefulSetDetails.ObjectMeta.Name, flags.QueueManagerNamespace)
	if err != nil {
		logger.Error(fmt.Sprintf("error while fetching StatefulSet events with name %s: %v", statefulSetDetails.ObjectMeta.Name, err))
	}

	return statefulSetList, statefulSetRevisionList, statefulSetEvents, nil
}
