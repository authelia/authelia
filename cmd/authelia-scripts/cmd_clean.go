package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Clean artifacts built and installed by authelia-scripts
func Clean(cobraCmd *cobra.Command, args []string) {
	fmt.Println("Removing `" + OutputDir + "` directory")
	err := os.RemoveAll(OutputDir)

	if err != nil {
		panic(err)
	}
}
