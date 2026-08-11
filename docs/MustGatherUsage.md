## MQ container inspector mustgather command

The MQ container inspector `mustgather` command collects diagnostic data from the specified queue manager running in a Kubernetes or OpenShift environment. This data is required when opening a case with IBM® MQ containers Support.

### MustGather usage examples

1. Gather diagnostics data using --qm-name:

`mq-container-inspector mustgather --qm-name <queue-manager-name> --qm-namespace <queue-manager-namespace>`

In this example, MQ container inspector uses the `app.kubernetes.io/instance: <queue-manager-name>` label to identify queue manager resources in the specified namespace. When deploying using the MQ Operator, this label is set to the queue manager custom resource `metadata.name`. When deploying using an mq sample helm chart, this label is set to the helm release name.

2. Gather diagnostics data using --pod-name:

`mq-container-inspector mustgather --pod-name <pod-name> --qm-namespace <queue-manager-namespace>`

In this example, MQ container inspector gets the Kubernetes pod with the specified `<pod-name>` and identifies any resources such as as services, statefulset or other queue manager pods which are connected to the specified pod. Using `--pod-name` is recommended if the queue manager is not deployed via the MQ operator or an mq sample helm chart.

### Kubernetes permissions required for mustgather

- The user requires "get" permissions for the resources listed in the [Example files gathered](#example-files-gathered). If the user doesn't have permissions for any resource, they will not be gathered, but the tool will continue to gather other resources where it does have permissions.
- The user requires "create" for "pods/exec" to run runmqras on the queue manager and to gather queue manager log files. Gathering these files can be disabled using the `--skip-exec` flag.

### full mustgather usage options

Below is the full usage options of the must gather tool. This is equivent to the output of the `--help` command

```
Usage of mustgather:
  --help
        show help message
  --kubeconfig
        path to the kubeconfig file. Ignored when running via mustgather image (default: ~/.kube/config)
  --operator-namespace
        namespace where the MQ operator is deployed (default: --qm-namespace)
  --output-dir
        directory where the must-gather output folder will be created. Ignored when running via mustgather image (default: current working directory)
  --pod-name
        name of a queue manager pod in the target queue manager instance (exactly one of --qm-name or --pod-name are required)
  --qm-container
        name of the container running MQ (default: qmgr)
  --qm-name
        QueueManager custom resource metadata.name (exactly one of --qm-name or --pod-name are required)
  --qm-namespace
        namespace where the queue manager is deployed (required)
  --pod-discovery
        enforce pod-based resource discovery instead of using the app.kubernetes.io/instance label when --pod-name is specified
  --skip-exec
        skip gathering diagnostic data that require container exec access, e.g. runmqras, webconsole logs (default: false)
  --skip-tar
        skip compressing the must-gather output into a tar.gz file (default: false)
```

### Example files gathered

#### Root Files
- runmqras-&lt;qm-name&gt;-ibm-mq-0.zip
- must-gather-logs.log

#### mq-operator/
- ibm-mq-operator-current-pod-log.txt
- ibm-mq-operator-previous-pod-log.txt
- ibm-mq-operator-deployment.yaml
- ibm-mq.&lt;version&gt;-csv.yaml

#### queue-managers/
- &lt;qm-name&gt;.yaml
- queue-manager-crd.yaml

#### routes/
- &lt;qm-name&gt;-ibm-mq-qm-routes-to-qm.yaml
- &lt;qm-name&gt;-ibm-mq-qm-routes.yaml
- &lt;qm-name&gt;-ibm-mq-web-routes-to-qm.yaml
- &lt;qm-name&gt;-ibm-mq-web-routes.yaml

#### services/
- &lt;qm-name&gt;-ibm-mq-metrics-service.yaml
- &lt;qm-name&gt;-ibm-mq-service.yaml

#### pods/
- &lt;qm-name&gt;-ibm-mq-0-current-pod-log.txt
- &lt;qm-name&gt;-ibm-mq-0-describe-log.txt
- &lt;qm-name&gt;-ibm-mq-0-pod-events.txt
- &lt;qm-name&gt;-ibm-mq-0.yaml
- pod-details.txt

#### statefulsets/
- &lt;qm-name&gt;-ibm-mq-statefulset-events.txt
- &lt;qm-name&gt;-ibm-mq-statefulset-revisions.yaml
- &lt;qm-name&gt;-ibm-mq-statefulset.yaml

#### pvcs/
- data-&lt;qm-name&gt;-ibm-mq-0-pvc.yaml

#### webconsole/
- web-&lt;qm-name&gt;-ibm-mq-0-console.log
- web-&lt;qm-name&gt;-ibm-mq-0-messages.log

#### configmaps/
- &lt;mqsc-configmap&gt;.yaml
- &lt;ini-configmap&gt;.yaml
