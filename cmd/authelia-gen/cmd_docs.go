package main

import (
	"github.com/spf13/cobra"
)

func newDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate docs",
		RunE:  docsRunE,
	}

	cmd.PersistentFlags().StringP("cwd", "C", "", "Sets the CWD for git commands")
	cmd.AddCommand(newDocsCLICmd(), newDocsDateCmd())

	return cmd
}

func docsRunE(cmd *cobra.Command, args []string) (err error) {
	for _, subCmd := range cmd.Commands() {
		switch {
		case subCmd.RunE != nil:
			if err = subCmd.RunE(subCmd, args); err != nil {
				return err
			}
		case subCmd.Run != nil:
			subCmd.Run(subCmd, args)
		}
	}

	return nil
}
