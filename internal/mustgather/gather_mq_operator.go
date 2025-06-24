package mustgather

import (
	"fmt"

	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/csv"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/deployment"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/kubeclient"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/namespace"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/pods"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func gatherMQOperatorToFiles(cfg *rest.Config, flags utils.MustGatherFlags) error {

	// build core client from config
	coreClient, err := kubeclient.BuildKubernetesClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building core client from config: %v", err)
	}

	// build dynamic client from config
	dynamicClient, err := kubeclient.BuildKubernetesDynamicClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building dynamic client from config: %v", err)
	}

	// operatorLabelSelector filters for the IBM MQ operator details using two labels:
	//   - "app.kubernetes.io/name=ibm-mq": matches IBM MQ resources.
	//   - "control-plane=controller-manager": matches operator resources.
	operatorLabelSelector := "app.kubernetes.io/name=ibm-mq,control-plane=controller-manager"

	operatorCSVLabelSelector := "app.kubernetes.io/name=ibm-mq,app.kubernetes.io/managed-by=olm"

	mqOperatorCSVFileNameFormat := "%s-csv.yaml"
	mqOperatorDeploymentFileNameFormat := "%s-deployment.yaml"
	mqOperatorPodLogsFileNameFormat := "%s-%s-pod-log.txt"

	// identify the operator namespace
	operatorNamespace, err := identifyOperatorNamespace(coreClient, dynamicClient, operatorLabelSelector, operatorCSVLabelSelector, flags)
	if err != nil {
		return err
	}
	fmt.Printf("mq-operator deployment found in %s namespace\n", operatorNamespace)

	// get the operator CSV details
	csvDetailsList, err := csv.GetOperatorCSVBySelector(dynamicClient, operatorCSVLabelSelector, operatorNamespace)
	if err == nil && csvDetailsList != nil {
		if err := csv.WriteCSVYamlsToFile(csvDetailsList.Items, mqOperatorCSVFileNameFormat, flags.OutputDir); err != nil {
			return err
		}
	} else {
		fmt.Printf("mq-operator CSV not found in namespace %q: %v\n", operatorNamespace, err)
	}

	// get the mq-operator deployment details
	deploymentList, err := deployment.GetDeploymentsBySelector(coreClient, operatorLabelSelector, operatorNamespace)
	if err != nil {
		return fmt.Errorf("error while fetching mq operator deployments with selector %s: %v", operatorLabelSelector, err)
	}

	// write the mq-deployments in their respective yamls
	if err := deployment.WriteDeploymentsToFile(deploymentList, mqOperatorDeploymentFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	// get the mq-operator pod logs
	podLogs, err := pods.GetPodLogsBySelector(coreClient, operatorLabelSelector, operatorNamespace, "manager")
	if err != nil {
		return fmt.Errorf("error while fetching mq operator pod logs with selector %s: %v", operatorLabelSelector, err)
	}

	// write the mq-operator pod logs in their respective log files
	if err := pods.WritePodLogsToFile(podLogs, mqOperatorPodLogsFileNameFormat, flags.OutputDir); err != nil {
		return err
	}

	return nil
}

func identifyOperatorNamespace(coreClient kubernetes.Interface, dynamicClient dynamic.Interface, operatorLabelSelector, operatorCSVLabelSelector string, flags utils.MustGatherFlags) (string, error) {

	if flags.OperatorNamespace != "" {
		// check if the operator-namespace is provided

		namespaceExists, err := namespace.DoesNamespacesExist(coreClient, flags.OperatorNamespace)
		if err != nil {
			return flags.OperatorNamespace, fmt.Errorf("error while checking if the namespace %v exist: %v", flags.OperatorNamespace, err)
		} else if !namespaceExists {
			return flags.OperatorNamespace, fmt.Errorf("provided mq-operator namespace %v was not found on the cluster", flags.OperatorNamespace)
		}

		return flags.OperatorNamespace, nil
	}

	// check in QueueManagerNamespace
	if foundMQOperatorCSV(dynamicClient, operatorCSVLabelSelector, flags.QueueManagerNamespace) {
		return flags.QueueManagerNamespace, nil
	}

	if foundMQOperatorDeployment(coreClient, operatorLabelSelector, flags.QueueManagerNamespace) {
		return flags.QueueManagerNamespace, nil
	}

	// check in GlobalOperatorNamespace exist, because it will only exist for OCP deployments
	if namespaceExist, _ := namespace.DoesNamespacesExist(coreClient, utils.GlobalOperatorNamespace); namespaceExist {

		if foundMQOperatorCSV(dynamicClient, operatorCSVLabelSelector, flags.QueueManagerNamespace) {
			return utils.GlobalOperatorNamespace, nil
		}

		if foundMQOperatorDeployment(coreClient, operatorLabelSelector, flags.QueueManagerNamespace) {
			return utils.GlobalOperatorNamespace, nil
		}

	}

	return "", fmt.Errorf("could not find MQ operator in any namespace (checked: %q, %q, %q)", flags.OperatorNamespace, flags.QueueManagerNamespace, utils.GlobalOperatorNamespace)

}

func foundMQOperatorDeployment(client kubernetes.Interface, selector, namespace string) bool {
	deploymentList, err := deployment.GetDeploymentsBySelector(client, selector, namespace)
	return err == nil && len(deploymentList) > 0
}

func foundMQOperatorCSV(client dynamic.Interface, selector, namespace string) bool {
	csvList, _ := csv.GetOperatorCSVBySelector(client, selector, namespace)
	return csvList != nil && len(csvList.Items) > 0
}
