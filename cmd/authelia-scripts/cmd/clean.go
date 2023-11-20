package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newCleanCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "clean",
		Short:   cmdCleanShort,
		Long:    cmdCleanLong,
		Example: cmdCleanExample,
		Args:    cobra.NoArgs,
		Run:     cmdCleanRun,

		DisableAutoGenTag: true,
	}

	return cmd
}

func cmdCleanRun(_ *cobra.Command, _ []string) {
	log.Debug("Removing `" + OutputDir + txtDirectoryTidle)
	err := os.RemoveAll(OutputDir)

	if err != nil {
		panic(err)
	}
}
