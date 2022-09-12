package main

import (
	"embed"

	"github.com/spf13/cobra"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	if err := newRootCmd().Execute(); err != nil {
		panic(err)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authelia-gen",
		Short: "Authelia's generator tooling",

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newAllCmd(), newCodeCmd(), newDocsCmd())

	return cmd
}
