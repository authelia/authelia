package main

import (
	"os"

	"github.com/clems4ever/authelia/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ServeCmd serve authelia with the provided configuration
func ServeCmd(cobraCmd *cobra.Command, args []string) {
	log.Infof("Running Authelia with config %s...", args[0])
	cmd := utils.CommandWithStdout(OutputDir+"/authelia", "-config", args[0])
	cmd.Env = append(os.Environ(), "PUBLIC_DIR=dist/public_html")
	utils.RunCommandUntilCtrlC(cmd)
}
