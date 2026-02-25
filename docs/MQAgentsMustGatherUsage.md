# MQ Agents Must Gather command

The MQ Agents `mustgather` command collects diagnostic data from the specified mq-agent deployment in an OpenShift environment. This data is required when opening a case for MQ Agents with IBM® Support.

## MQ Agent Must Gather usage examples

1. Essential command (mandatory flags only)
```sh
mq-container-inspector mq-agent --agent-release-name <name-of-the-mq-agent> --namespace <namespace-where-mq-agent-is-deployed>
```

The above command consists of the mandatory flags required to run the must-gather tool for mq-agents. The `--agent-release-name` flag points to the MQ-Agent deployment name you are running the must-gather against, while the `--namespace` flag points to the namespace where your deployment is present in. This command is the bare minimum required to run the mq-agent must-gather tool, where the `kubeconfig` will be defaulted to the `~/.kube/config` file and `output-dir` to the current `pwd` from where you are running the tool.

2. Specifying the output-directory
```sh
mq-container-inspector mq-agent --agent-release-name <name-of-the-mq-agent> --namespace <namespace-where-mq-agent-is-deployed> --output-dir <output-dir-where-must-gathers-is-saved>
```

In the above command the `--output-dir` flag is used to specify the location where you want the must gathers to be collected in. 
> [!NOTE]
> If you are running the tool off the mustgather image, the above flag will be ignored even if you specify it and will default to the current working directory.

3. Specifying the kubeconfig
```sh
mq-container-inspector mq-agent --agent-release-name <name-of-the-mq-agent> --namespace <namespace-where-mq-agent-is-deployed> --kubeconfig <kubeconfig-file-location-you-want-to-specify>
```

In the above command the `--kubeconfig` flag is used to specify the location of your custom kubeconfig file which you want the tool to use.
> [!NOTE]
> If you are running the tool off the mustgather image, the above flag will be ignored even if you specify it and will default to the `~/.kube/config` file.

4. Skipping the tar compression of the must-gathers
```sh
mq-container-inspector mq-agent --agent-release-name <name-of-the-mq-agent> --namespace <namespace-where-mq-agent-is-deployed> --skip-tar
```

In the above command `--skip-tar` flag is used to skip the tar compression of the agent must-gather. By default, this flag is set to `false`.

## OpenShift permissions required for the mq-agent mustgather

- The user requires "get" permissions for all the resources listed in the [Example files gathered](#example-files-gathered). If the user doesn't have permissions for any of the resources, the tool will fail with an error message mentioning the resource they do not have access for.

## full mq-agent mustgather usage options

Below is the full usage options for the must gather tool. This is equivalent to the output of the `--help` command
```
Usage of mq-agent:
--agent-release-name
        name of the mq-agent release(required)
--help
        show help message
--kubeconfig
        path to the kubeconfig file. Ignored when running via mustgather image (default: ~/.kube/config)
--namespace
        namespace where the mq-agent is deployed(required)
--output-dir
        directory where the must-gather output folder will be created. Ignored when running via mustgather image (default: current working directory)
--skip-tar
        skip compressing the mq-agent must-gather output into a tar.gz file (default: false)
```

## Example files gathered

### Root Files
- mq-agent-must-gather-logs.log

### configmap/
- `<configmap-name>`.yaml

### deployment/
- `ibm-mq-agent-<agent-release-name>-deployment`.yaml
- `ibm-mq-agent-<agent-release-name>-deployment-events`.txt

### networkpolicy/
- `ibm-mq-agent-<agent-release-name>-network-policy`.yaml

### pods/
- `[agent]-<pod-name>-current-pod-log`.txt
- `[agent]-<pod-name>-previous-pod-log`.txt
- `[mcp-server]-<pod-name>-current-pod-log`.txt
- `[mcp-server]-<pod-name>-previous-pod-log`.txt
- `<pod-name>-describe-log`.txt
- `<pod-name>`.yaml
- pod-details.txt
- `<pod-name>`-pod-events.txt

### replicaset/
- `<replicaset-name>`.yaml
- `<replicaset-name>-replicaset-events`.txt

### route/
- `ibm-mq-agent-<agent-release-name>-route`.yaml

### service/
- `ibm-mq-agent-<agent-release-name>-service`.yaml

### service-account/
- `ibm-mq-agent-<agent-release-name>-service-account`.yaml