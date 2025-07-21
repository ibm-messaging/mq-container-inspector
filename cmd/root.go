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

	cmd "github.ibm.com/mq-cloudpak/mq-inspector/cmd/mustgather"
	"github.ibm.com/mq-cloudpak/mq-inspector/pkg/utils"
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
		err := cmd.MustGather(args[1:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("MustGathers collected")
	case utils.Version:
		printVersion()
	default:
		printHelp()
	}

}

func printHelp() {
	// TODO: add full usage info
	fmt.Println(("Usage info"))
}

// TODO: get version from build flag so we don't need to keep updating this function
func printVersion() {
	fmt.Println("mq-inspector version v1.0.0")
}
