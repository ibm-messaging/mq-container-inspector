package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/statefulset"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/kubernetes"
)

func gatherStatefulSetToFiles(coreClient kubernetes.Interface, flags utils.MustGatherFlags) error {

	statefulSetLabelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", flags.QueueManagerName)

	// get StatefulSet details by selector
	statefulSetList, err := statefulset.GetStatefulSetDetailsBySelector(coreClient, statefulSetLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching StatefulSets with selector %s: %v", statefulSetLabelSelector, err)
	}

	// write the StatefulSet details to their yamls
	statefulSetDetailsFileNameFormat := "%s-statefulset.yaml"
	if err := statefulset.WriteStatefulSetYamlsToFile(statefulSetList, statefulSetDetailsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// get StatefulSet revisions by selector
	statefulSetRevisionList, err := statefulset.GetStatefulSetRevisionsBySelector(coreClient, statefulSetLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching StatefulSet revisions with selector %s: %v", statefulSetLabelSelector, err)
	}

	// write the StatefulSet revisions to their yamls
	statefulSetRevisionsFileNameFormat := "%s-statefulset-revisions.yaml"
	if err := statefulset.WriteStatefulSetRevisionYamlsToFile(statefulSetRevisionList, statefulSetRevisionsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// get StatefulSet events by selector
	statefulSetEvents, err := statefulset.GetStatefulSetEventsBySelector(coreClient, statefulSetLabelSelector, flags.QueueManagerNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching StatefulSet events with selector %s: %v", statefulSetLabelSelector, err)
	}

	// write the StatefulSet events to their files
	statefulSetEventsFileNameFormat := "%s-statefulset-events.txt"
	if err := statefulset.WriteStatefulSetEventsToFile(statefulSetEvents, statefulSetEventsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	return nil

}
