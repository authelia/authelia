package main

import (
	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		panic(err)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authelia-gen",
		Short: "Authelia's generator tooling",
	}

	cmd.PersistentFlags().StringP("cwd", "C", "", "Sets the CWD for git commands")

	cmd.AddCommand(newAllCmd(), newCodeCmd(), newDocsCmd(), newGitHubCmd(), newLocalesCmd(), newCommitLintCmd())

	return cmd
}
