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

	cmd.AddCommand(newDocsCLICmd(), newDocsDataCmd(), newDocsDateCmd(), newDocsSEOCmd(), newDocsJSONSchemaCmd(), newDocsManageCmd())

	return cmd
}

func newDocsManageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseManage,
		Short: "Generate Managed docs",

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newADRCmd())

	return cmd
}
