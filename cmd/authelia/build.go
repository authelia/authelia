package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Show the build information of Authelia",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Authelia version %s, build %s\n", BuildTag, BuildCommit)
	},
}
