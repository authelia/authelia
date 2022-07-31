package main

import (
	"github.com/spf13/cobra"
)

func newDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate docs",
		RunE:  rootSubCommandsRunE,
	}

	cmd.AddCommand(newDocsCLICmd(), newDocsDateCmd())

	return cmd
}
