## PVCTool Usage

The MQ container inspector PVC tool creates pod(s) attached to the queue manager PVCs. When queue manager pods are available, the PVC-inspector pods are created using information from the existing queue manager pods. When the queue manager pods are not available, the `--qm-image` fallback can be used to create PVC-inspector pods from the remaining PVCs.

You can then exec into the PVC-inspector pods to investigate the PVCs manually, or run the tool with the `--runmqras` flag to run the `runmqras` must gather on the pods.

### PVCTool usage examples

1. Create pvctool pods using `--qm-name`:

```bash
mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <queue-manager-namespace>
```

This command creates pods attached to the queue manager PVCs. You can exec into these pods to manually inspect the PVCs in a Kubernetes or OpenShift environment.

**Note:** If you exec into a pvctool pod, any changes to the PVC will affect the corresponding queue manager and could result in a corrupted queue manager or lost messages.

To delete the pvctool pods, run the same command with the `--cleanup` flag:

```bash
mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <namespace> --cleanup
```

2. Perform a `runmqras` must gather using the PVC tool:

```bash
mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <queue-manager-namespace> --runmqras --cleanup
```

This command creates pods attached to the queue manager PVCs. It then execs into the pvctool pods, runs `runmqras`, and deletes the pvctool pods.

By default, the output is saved inside a `PVC_Inspector_<timestamp>` folder in the current working directory.

3. Create pvctool pods when the queue manager pods do not exist:

```bash
mq-container-inspector pvctool \
  --qm-name <queue-manager-name> \
  --qm-namespace <queue-manager-namespace> \
  --qm-image <queue-manager-image>
```

The `--qm-image` option enables the PVC tool to create PVC-inspector pods when the original queue manager pods are no longer available.

The value can be either a QueueManager version:

```bash
--qm-image 10.0.0.0-r2
```

or a full container image reference:

```bash
--qm-image <registry>/<repository>:<tag>
```

When a version is supplied, the tool resolves it to the default IBM MQ QueueManager image.

If a QueueManager custom resource exists, the tool also uses the custom resource to validate the QueueManager version and determine whether the QueueManager uses ephemeral storage.

If `--qm-image` is not specified and the QueueManager pods do not exist, the tool attempts to obtain the QueueManager version from the QueueManager custom resource.

### Using `--qm-image` with MQ Operator deployments

For an MQ Operator deployment, the PVC tool supports the fallback when the QueueManager pods no longer exist but the QueueManager PVCs are still present.

If the QueueManager custom resource still exists:

* `--qm-image` can be supplied explicitly.
* If `--qm-image` is not supplied, the tool derives the QueueManager image version from the QueueManager custom resource.
* The QueueManager custom resource is used to determine whether the QueueManager is configured with ephemeral storage.
* If a version supplied through `--qm-image` differs from the QueueManager custom resource version, the QueueManager custom resource version is used when the default IBM MQ image is being resolved.
* If a full image reference is supplied, the supplied image is preserved.

If the QueueManager custom resource no longer exists:

* `--qm-image` must be supplied because the QueueManager image can no longer be derived from the custom resource.
* The remaining PVCs are discovered using the QueueManager instance label and supported MQ Operator PVC naming conventions.

The PVC tool currently expects the PVCs to contain:

```text
app.kubernetes.io/instance=<queue-manager-name>
```

and to use the supported MQ Operator PVC naming conventions.

This includes PVCs used by SingleInstance, NativeHA, and MultiInstance deployments.

The fallback cannot inspect a QueueManager configured with ephemeral storage because there are no persistent QueueManager PVCs to attach to the PVC-inspector pods.

PVCs whose labels or names have been customized such that they cannot be identified by the discovery logic are not automatically inspected.

### Using `--qm-image` with non-Operator deployments

For non-Operator deployments, `--qm-image` can be used when the QueueManager pods no longer exist but the QueueManager PVCs remain.

For example:

```bash
mq-container-inspector pvctool \
  --qm-name <instance-name> \
  --qm-namespace <namespace> \
  --qm-image <queue-manager-image>
```

The PVC discovery logic expects:

* `app.kubernetes.io/instance=<instance-name>` on the PVCs.
* Supported default PVC names using the `qm`, `data`, and `log` naming conventions.

These PVCs are used to identify the QueueManager data, persisted data, and recovery log storage that should be mounted into the PVC-inspector pods.

The fallback might not be able to identify the required PVCs if the PVC labels or naming conventions have been customized.

### Kubernetes permissions required for pvctool

* The user requires `get` permissions for `pods` and `configmaps`.
* The user requires `create` permissions for `pods` and `configmaps`, because these resources are created when the tool runs.
* The user requires `create` permission for `pods/exec` to run `runmqras` on the pods created by pvctool.

## Full pvctool usage options

Below are the full usage options of the PVC tool. This is equivalent to the output of the `--help` command.

```text
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

  --qm-image
        Queue manager container image used for the PVC inspector pod if no queue manager pods exist. Specify either a queue manager container V.R.M.F-release version or a fully qualified container image reference. Requires --qm-name.

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

If you use the `--runmqras` flag, the pvctool command creates a folder called `PVC_Inspector_<timestamp>` in the `--output-dir`, which defaults to the current working directory.

This contains a [runmqras](https://www.ibm.com/docs/en/ibm-mq/9.4.x?topic=reference-runmqras-collect-mq-troubleshooting-information) tarball for each queue manager and a log file for the tool.

Unless `--skip-tar` is specified, the tool also creates a compressed tarball of the `PVC_Inspector_<timestamp>` folder, which can be uploaded to the support case.
