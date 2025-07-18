# IBM MQ Inspector
Test1
## Overview

A collection of tools for interacting with IBM® MQ queue managers running in containers.

This repository currently includes the MQ MustGather tool, which simplifies the collection of diagnostic information from a running queue manager. Tools are provided as CLI commands, built from a Go binary, and are designed to be lightweight, container-friendly, and easy to integrate into automated environments.

Additional tools may be added over time to support broader inspection and interaction use cases.

## Build

To build the MQ Inspector tools, simply run:

`go build`

This will produce a single binary containing all available CLI tools. No additional dependencies or setup steps are required. 

## Usage

After building the binary, you can run the MustGather tool with the following command:

`./mq-inspector mustgather --qm-name <queue manager name> --qm-namespace <queue manager namespace>`

This command collects diagnostic data from the specified queue manager running in a Kubernetes or OpenShift environment.

By default, the output is saved inside a `Must_Gather_<timestamp>` folder in the current working directory.

## Issues and contributions

For issues relating specifically to the MQ Inspector tool, please use the [GitHub issue tracker](https://github.com/ibm-messaging/mq-inspector/issues). Pull requests are not currently accepted.

## License

This tool is licensed under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

## Copyright

© Copyright IBM Corporation 2025
