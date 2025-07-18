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
