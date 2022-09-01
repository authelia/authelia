package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newCICmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "ci",
		Short:   cmdCIShort,
		Long:    cmdCILong,
		Example: cmdCIExample,
		Args:    cobra.NoArgs,
		Run:     cmdCIRun,

		DisableAutoGenTag: true,
	}

	return cmd
}

func cmdCIRun(cmd *cobra.Command, _ []string) {
	log.Info("=====> Build stage <=====")

	if buildkite, _ := cmd.Flags().GetBool("buildkite"); buildkite {
		if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "--buildkite", "build").Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "build").Run(); err != nil {
			log.Fatal(err)
		}
	}

	log.Info("=====> Unit testing stage <=====")

	if err := utils.CommandWithStdout("authelia-scripts", "--log-level", "debug", "unittest").Run(); err != nil {
		log.Fatal(err)
	}
}
