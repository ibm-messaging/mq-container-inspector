## PVCTool Usage

The MQ container inspector PVC tool creates pod(s) the the same nodes as the queue manager pod(s) attached to the queue manager PVCs. You can then exec into the pvc pods to investigate the pvcs manually or run the tool with the `--runmqras` flag to run the runmqras must gather on the pod.

### PVCTool usage examples

1. Create pvctool pods using --qm-name:

`mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <queue-manager-namespace>`

This command creates pods attached to the queue manager PVCs. You can exec into these pods to manually inspect the PVCs in a Kubernetes or OpenShift environment. 

**Note**: If you exec into a pvctool pod, any changes to the PVC will affect the corresponding queue manager and could result in a corrupted queue manager or lost messages. 

To delete the pvctool pods, run the same command with the --cleanup flag:

`mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <namespace> --cleanup`

2. Perform a runmqras must gather using the pvc tool:

`mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <queue-manager-namespace> --runmqras --cleanup` 

This command creates pods attached to the queue manager PVCs. It then execs into the pvctool pod, runs runmqras and deletes the pvctool pods.

By default, the output is saved inside a `PVC_Inspector_<timestamp>` folder in the current working directory.

### Kubernetes permissions required for pvctool

- The user requires "get" permissions for "pods" and "configmaps".
- The user requires "create" permissions for "pods" and "configmaps" and will create these resources when it is run. 
- The user requires "create" for "pods/exec" to run runmqras on the pods created by pvctool.


## full pvctool usage options

Below is the full usage options of the must gather tool. This is equivent to the output of the `--help` command

```
Usage of pvctool:
  --cleanup
        delete the pvc-inspector pods at the end of the tool run (default: false)
  --dry-run
        run without creating the pvc-inspector pods (default: false)
  --help
        show help message
  --kubeconfig
        path to the kubeconfig file. Ignored when running via mustgather image. (default: ~/.kube/config)
  --output-dir
        directory where the must-gather output folder will be created. Ignored when running via mustgather image. (default: current working directory)
  --pod-name
        name of a queue manager pod in the target queue manager instance (exactly one of --qm-name or pod-name are required)
  --qm-name
        QueueManager custom resource metadata.name (exactly one of --qm-name or pod-name are required)
  --qm-namespace
        namespace where the queue manager is deployed (required)
  --runmqras
        execute runmqras on the pvc-inspector pods (default: false)
  --skip-tar
        skip compressing the pvctool output into a tar.gz file (default: false)
```

### Example tool output

If you use the `--runmqras` flag, the pvctool command will create a folder called `PVC_Inspector_<timestamp>` in the `--output-dir` (defaults to current working directory). This will contain a [runmqras](https://www.ibm.com/docs/en/ibm-mq/9.4.x?topic=reference-runmqras-collect-mq-troubleshooting-information) tarball for each queue manager and a logfile for the tool. 

The tool will also create a tarball of the PVC_Inspector_<timestamp> folder which should be uploaded to the support case.