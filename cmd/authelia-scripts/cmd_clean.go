package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Clean artifacts built and installed by authelia-scripts.
func Clean(cobraCmd *cobra.Command, args []string) {
	log.Debug("Removing `" + OutputDir + "` directory")
	err := os.RemoveAll(OutputDir)

	if err != nil {
		panic(err)
	}
}
