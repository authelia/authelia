package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/utils"
)

// RunCI run the CI scripts.
func RunCI(cmd *cobra.Command, args []string) {
	log.Info("=====> Build stage <=====")
	if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "build").Run(); err != nil {
		log.Fatal(err)
	}

	log.Info("=====> Unit testing stage <=====")
	if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "unittest").Run(); err != nil {
		log.Fatal(err)
	}
}
