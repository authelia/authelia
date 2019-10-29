package main

import (
	"os"

	"github.com/spf13/cobra"
)

// ServeCmd serve authelia with the provided configuration
func ServeCmd(cobraCmd *cobra.Command, args []string) {
	cmd := CommandWithStdout(OutputDir+"/authelia", "-config", args[0])
	cmd.Env = append(os.Environ(), "PUBLIC_DIR=dist/public_html")
	RunCommandUntilCtrlC(cmd)
}
