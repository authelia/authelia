package commands

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newBuildInfoCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "build-info",
		Short:   cmdAutheliaBuildInfoShort,
		Long:    cmdAutheliaBuildInfoLong,
		Example: cmdAutheliaBuildInfoExample,
		RunE:    ctx.BuildInfoRunE,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}

// BuildInfoRunE is the RunE for the authelia build-info command.
func (ctx *CmdCtx) BuildInfoRunE(cmd *cobra.Command, _ []string) (err error) {
	var verbose bool
	if verbose, err = cmd.Flags().GetBool("verbose"); err != nil {
		return err
	}

	return runBuildInfo(cmd.OutOrStdout(), verbose)
}

func runBuildInfo(w io.Writer, verbose bool) (err error) {
	var (
		info *debug.BuildInfo
		ok   bool
	)

	if info, ok = debug.ReadBuildInfo(); !ok {
		return fmt.Errorf("failed to read build info")
	}

	_, _ = fmt.Fprintf(w, fmtAutheliaBuild, utils.BuildTag, utils.BuildState, utils.BuildBranch, utils.BuildCommit,
		utils.BuildNumber, runtime.GOOS, runtime.GOARCH, runtime.Compiler, utils.BuildDate, utils.Dev, utils.BuildExtra)

	_, _ = fmt.Fprintf(w, "\n"+fmtAutheliaBuildGo, info.GoVersion, info.Main.Path, info.Path)

	if verbose {
		if len(info.Settings) != 0 {
			_, _ = fmt.Fprintf(w, "    Settings:\n")

			for _, setting := range info.Settings {
				_, _ = fmt.Fprintf(w, "        %s: %s\n", setting.Key, setting.Value)
			}
		}

		if len(info.Deps) != 0 {
			_, _ = fmt.Fprintf(w, "    Dependencies:\n")

			for _, dep := range info.Deps {
				_, _ = fmt.Fprintf(w, "        %s@%s (%s)\n", dep.Path, dep.Version, dep.Sum)
			}
		}
	}

	return nil
}
