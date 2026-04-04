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
	buildkite, _ := cmd.Flags().GetBool("buildkite")

	args := []string{"--log-level", "debug"}

	if buildkite {
		args = append(args, "--buildkite")
	}

	log.Info("=====> Build stage <=====")

	if err := utils.CommandWithStdout("authelia-scripts", append(args, "build")...).Run(); err != nil {
		log.Fatal(err)
	}

	log.Info("=====> Unit testing stage <=====")

	if err := utils.CommandWithStdout("authelia-scripts", append(args, "unittest")...).Run(); err != nil {
		log.Fatal(err)
	}
}
