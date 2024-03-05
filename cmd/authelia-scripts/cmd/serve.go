package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newServeCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "serve [config]",
		Short:   cmdServeShort,
		Long:    cmdServeLong,
		Example: cmdServeExample,
		Args:    cobra.MinimumNArgs(1),
		Run:     cmdServeRun,

		DisableAutoGenTag: true,
	}

	return cmd
}

func cmdServeRun(_ *cobra.Command, args []string) {
	log.Infof("Running Authelia with config %s...", args[0])
	execCmd := utils.CommandWithStdout(OutputDir+pathAuthelia, "--config", args[0])
	utils.RunCommandUntilCtrlC(execCmd)
}
