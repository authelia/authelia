package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewRootCmd returns the root authelia-scripts cmd.
func NewRootCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "authelia-scripts",
		Short:   cmdRootShort,
		Long:    cmdRootLong,
		Example: cmdRootExample,

		DisableAutoGenTag: true,
	}

	cmd.PersistentFlags().Bool("buildkite", false, "Set CI flag for Buildkite")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level for the command")

	cmd.AddCommand(newBootstrapCmd(), newBuildCmd(), newCleanCmd(), newCICmd(), newDockerCmd(), newServeCmd(), newSuitesCmd(), newUnitTestCmd(), newXFlagsCmd())

	cobra.OnInitialize(cmdRootInit)

	return cmd
}

func cmdRootInit() {
	log.SetLevel(levelStringToLevel(logLevel))
}
