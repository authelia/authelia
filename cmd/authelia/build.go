package main

import (
	"fmt"
	"github.com/authelia/authelia/internal/utils"
	"github.com/spf13/cobra"
	"runtime"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Show the build information of Authelia",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(fmtAutheliaBuild, utils.BuildTag, utils.BuildState, utils.BuildBranch, utils.BuildCommit,
			utils.BuildNumber, runtime.GOOS, runtime.GOARCH, utils.BuildDate, utils.BuildExtra)
	},
}
