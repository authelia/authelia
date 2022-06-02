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
		Use: "authelia-gen",
	}

	cmd.AddCommand(newAllCmd(), newCodeCmd(), newDocsCmd())

	return cmd
}
