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

	cmd.AddCommand(newAllCmd(), newCodeCmd(), newDocsCmd())

	return cmd
}
