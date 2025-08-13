package daemonset

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// WriteDaemonSetsToFile writes each daemonset details to a separate YAML file.
// Parameters:
//   - daemonSetList: the list of daemonsets whose details will be written.
//   - fileNameFormat: the format string used to name each file, must contain one "%s", which will be replaced by the daemonset name.
//   - outputDir:      the directory in which the YAML files will be created.
func WriteDaemonSetsToFile(daemonSetList []appsv1.DaemonSet, fileNameFormat, outputDir string) error {

	for _, daemonSet := range daemonSetList {

		fileName := utils.FormatFilePath(outputDir, fileNameFormat, daemonSet.ObjectMeta.Name)

		data, err := yaml.Marshal(daemonSet)
		if err != nil {
			return fmt.Errorf("error while marshalling yaml for daemonset %s: %v", daemonSet.ObjectMeta.Name, err)
		}

		if err := os.WriteFile(fileName, data, 0660); err != nil {
			return fmt.Errorf("error while writing %s daemonset data in the yaml file: %v", daemonSet.ObjectMeta.Name, err)
		}

	}

	return nil

}

// WriteDaemonSetEventsToFile writes each daemonset events to a separate file.
// Parameters:
//   - daemonsetEventsEventMap: the map of daemonset names to their events, which will be written
//   - fileNameFormat: the format string used to name each file, must contain one "%s", which will be replaced by the daemonset name.
//   - outputDir:      the directory in which the event files will be created.
func WriteDaemonSetEventsToFile(daemonsetEventsEventMap map[string][]corev1.Event, fileNameFormat, outputDir string) error {

	for daemonSetName, daemonSetEvents := range daemonsetEventsEventMap {

		if len(daemonSetEvents) > 0 {

			fileName := utils.FormatFilePath(outputDir, fileNameFormat, daemonSetName)

			file, err := os.Create(fileName)
			if err != nil {
				return fmt.Errorf("error creating %s daemonset event file %s: %v", daemonSetName, fileName, err)
			}

			// tabwriter will handle dynamic spacing
			writer := tabwriter.NewWriter(file, 0, 8, 2, ' ', 0)

			// write the header
			fmt.Fprintf(writer, "LAST SEEN\tTYPE\tREASON\tKIND\tNAME\tMESSAGE\n")

			for _, event := range daemonSetEvents {

				var lastSeen string
				if !event.LastTimestamp.IsZero() {
					lastSeen = event.LastTimestamp.Format(time.RFC3339)
				}

				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
					lastSeen,
					event.Type,
					event.Reason,
					event.InvolvedObject.Kind,
					event.InvolvedObject.Name,
					event.Message,
				)
			}

			writer.Flush()
			file.Close()

		}
	}

	return nil

}
