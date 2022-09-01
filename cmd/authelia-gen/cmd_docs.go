package main

import (
	"github.com/spf13/cobra"
)

func newDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseDocs,
		Short: "Generate docs",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newDocsCLICmd(), newDocsDateCmd())

	return cmd
}
