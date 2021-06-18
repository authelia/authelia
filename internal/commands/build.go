package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/utils"
)

func newBuildCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:               "build",
		Short:             "Show the build information of Authelia",
		RunE:              runBuildCmdE,
		PersistentPreRunE: nil,
	}

	return cmd
}

func runBuildCmdE(_ *cobra.Command, _ []string) (err error) {
	_, err = fmt.Printf(fmtAutheliaBuild, utils.BuildTag, utils.BuildState, utils.BuildBranch, utils.BuildCommit,
		utils.BuildNumber, runtime.GOOS, runtime.GOARCH, utils.BuildDate, utils.BuildExtra)

	return err
}
