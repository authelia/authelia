package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

// ServeCmd serve Authelia with the provided configuration.
func ServeCmd(cmd *cobra.Command, args []string) {
	log.Infof("Running Authelia with config %s...", args[0])
	execCmd := utils.CommandWithStdout(OutputDir+"/authelia", "--config", args[0])
	utils.RunCommandUntilCtrlC(execCmd)
}
