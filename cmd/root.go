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
package cmd

import (
	"fmt"
	"os"

	"github.ibm.com/mq-cloudpak/mq-container-inspector/cmd/mustgather"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/cmd/pvcinspectortool"
	"github.ibm.com/mq-cloudpak/mq-container-inspector/pkg/utils"
)

func Execute() {
	// Skipping first arg as we know it is the path of the executable
	args := os.Args[1:]
	if len(args) < 1 {
		printHelp()
		os.Exit(1)
	}
	switch args[0] {
	case utils.MustGather:
		fmt.Println("Starting Must-Gather tool")
		err := mustgather.MustGather(args[1:])
		if err != nil {
			fmt.Printf("MustGather FAILED: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("MustGathers collected")
	case utils.PVCInspector:
		fmt.Println("Starting PVC-Inspector tool")
		err := pvcinspectortool.PVCIncpectorTool(args[1:])
		if err != nil {
			fmt.Printf("PVC-Inspector FAILED: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("PVC data collected")
	case utils.Version:
		printVersion()
	default:
		printHelp()
	}

}

func printHelp() {

	fmt.Println(`
mq-container-inspector: A CLI tool for inspecting MQ resources

Usage:
  mq-container-inspector <command> [flags]

Available Commands:
  mustgather     Collect diagnostic data for a queue manager
  pvctool        Inspect PVCs associated with a queue manager
  version        Print the version of mq-container-inspector

Examples:
  mq-container-inspector mustgather --qm-name <queue-manager-name> --qm-namespace <namespace>
  mq-container-inspector pvctool --qm-name <queue-manager-name> --qm-namespace <namespace>

Use "mq-container-inspector <command> --help" for more information about a command.`)

}

func printVersion() {
	fmt.Println("mq-container-inspector version: 1.0.0")
}
