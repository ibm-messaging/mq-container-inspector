# IBM MQ container inspector

## Overview

A collection of tools for interacting with IBM® MQ queue managers running in containers.

The following tools are currently available:

 - [MustGather](#mustgather-usage) - Collect diagnostic information required when opening a support case with IBM® MQ containers Support.
 - [PVCTool](#pvctool-usage) - Provides access to the files on a queue manager PVC where a remote shell cannot be established to the queue manager pod. This might be because the pod is in an Error or CrashLoopBackOff state.

## Install Instructions

To use MQ container inspector, you must install it via one of the following options: 

- [Build Binary](#build-binary): Build the MQ container inspector binary from the source code
- [Download Binary](#download-binary): Download the prebuilt MQ container inspector Binary
- [Prebuilt MQ container inspector image](#prebuilt-mq-container-inspector-image): Use MQ container inspector via the CPFS MustGather Image

### Build Binary

To build and install the MQ container inspector` tool, follow these steps:  

1. Clone MQ container inspector GitHub repository, for example:  

```bash
git clone https://github.com/ibm-messaging/mq-container-inspector.git 
cd mq-container-inspector
```

2. Build MQ container inspector:

```bash
go build
```

3. Add mq-container-inspector to your path, for example:

```
sudo mv mq-container-inspector /usr/local/bin
```

4. Verify the build was successful:  

`mq-container-inspector version`

### Download Binary

#### Linux

1. Download the latest MQ container inspector tarball and signature for your architecture from the [releases page](https://github.com/ibm-messaging/mq-container-inspector/releases)
2. Verify the tarball signature by following the instructions in [VerifyBinary.md](https://github.com/ibm-messaging/mq-container-inspector/tree/main/docs/VerifyBinary.md)
3. Extract the tarball (e.g. for 1.0.0_mq-container-inspector_linux-amd64.tar.gz):  
`tar -xvzf ~/Downloads/1.0.0_mq-container-inspector_linux-amd64.tar.gz`
4. Add mq-container-inspector to your path, for example:

```
sudo mv 1.0.0_mq-container-inspector_linux-amd64 /usr/local/bin/mq-container-inspector
```
5. Verify the Installation:  
`mq-container-inspector version`

#### MacOS
- Requires [Homebrew](https://brew.sh/)
- Requires [Go](https://go.dev/)

An MQ container inspector tarball is not available for MacOS, however the binary is available for install via brew.

1. Add the MQ homebrew tap:  
`brew tap ibm-messaging/ibmmq`

2. Install MQ container inspector:  
`brew install ibm-messaging/ibmmq/mqcontainerinspector`

3. Verify the Installation:  
`mq-container-inspector version`

#### Windows

1. Download the latest MQ container inspector zip file and signature from the [releases page](https://github.com/ibm-messaging/mq-container-inspector/releases)
2. Verify the zip file signature by following the instructions in [VerifyBinary.md](https://github.com/ibm-messaging/mq-container-inspector/tree/main/docs/VerifyBinary.md)
3. Unzip the Tool, for example:
- Locate the downloaded .zip file (e.g. 1.0.0_mq-container-inspector_windows-amd64.zip)
- Right-click the file and choose Extract All...
- Choose a destination folder and click Extract
4. Add the executable to your system path:  
`setx PATH "%PATH%;C:\<Path-to-executable>\mq-container-inspector" /M`
5. Verify the Installation:  
`mq-container-inspector version`

### Pre-built MQ container inspector image

The MQ container inspector binary is included in the IBM Cloud Pak foundational services MustGather image (icr.io/cpopen/cpfs/must-gather:latest). For usage instructions, see the [MQ Container troubleshooting docs](https://www.ibm.com/docs/en/ibm-mq/9.4.x?topic=troubleshooting-problems-mq-operator)

## Mustgather Usage

After the binary has been built or installed, you can run the MustGather tool with the following command:

`mq-container-inspector mustgather --qm-name <queue-manager-name> --qm-namespace <queue-manager-namespace>`

This command collects diagnostic data from the specified queue manager running in a Kubernetes or OpenShift environment.

By default, the output is saved inside a `Must_Gather_<timestamp>` folder in the current working directory.

For full usage and output, see the [Mustgather Usage Guide](https://github.com/ibm-messaging/mq-container-inspector/tree/main/docs/MustGatherUsage.md)

## PVCtool Usage

After the binary has been built or installed, you can use the PVC tool to perform a [runmqras](https://www.ibm.com/docs/en/ibm-mq/9.4.x?topic=reference-runmqras-collect-mq-troubleshooting-information) must gather against a failing queue manager, with the following command:

`mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <queue-manager-namespace> --runmqras --cleanup` 

This command creates kubernetes pods attached to the queue manager PVCs. It then execs into the pvctool pod, runs runmqras and deletes the pods.

By default, the output is saved inside a `PVC_Inspector_<timestamp>` folder in the current working directory.

For full usage and output, see the [PVCTool Usage Guide](https://github.com/ibm-messaging/mq-container-inspector/tree/main/docs/PVCToolUsage.md)

## Issues and contributions

For issues relating specifically to the MQ container inspector tool, please use the [GitHub issue tracker](https://github.com/ibm-messaging/mq-container-inspector/issues). Pull requests are not currently accepted.

## License

This tool is licensed under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

## Copyright

© Copyright IBM Corporation 2025
