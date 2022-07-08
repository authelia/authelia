package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newBuildInfoCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "build-info",
		Short:   cmdAutheliaBuildInfoShort,
		Long:    cmdAutheliaBuildInfoLong,
		Example: cmdAutheliaBuildInfoExample,
		RunE:    cmdBuildInfoRunE,
		Args:    cobra.NoArgs,
	}

	return cmd
}

func cmdBuildInfoRunE(_ *cobra.Command, _ []string) (err error) {
	_, err = fmt.Printf(fmtAutheliaBuild, utils.BuildTag, utils.BuildState, utils.BuildBranch, utils.BuildCommit,
		utils.BuildNumber, runtime.GOOS, runtime.GOARCH, utils.BuildDate, utils.BuildExtra)

	return err
}
