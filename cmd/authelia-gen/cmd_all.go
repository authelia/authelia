package main

import (
	"github.com/spf13/cobra"
)

func newAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Run all generators with default options",
		RunE:  allRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func allRunE(cmd *cobra.Command, args []string) (err error) {
	for _, subCmd := range cmd.Parent().Commands() {
		if subCmd == cmd || subCmd.Use == "completion" || subCmd.Use == "help [command]" {
			continue
		}

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
