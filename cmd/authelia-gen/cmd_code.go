package main

import (
	"github.com/spf13/cobra"
)

func newCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code",
		Short: "Generate code",
		RunE:  codeRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newCodeKeysCmd())

	return cmd
}

func codeRunE(cmd *cobra.Command, args []string) (err error) {
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
