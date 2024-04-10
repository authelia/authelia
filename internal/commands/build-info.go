package commands

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

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
	var (
		info *debug.BuildInfo
		ok   bool
	)

	if info, ok = debug.ReadBuildInfo(); !ok {
		return fmt.Errorf("failed to read build info")
	}

	b := &strings.Builder{}

	b.WriteString(fmt.Sprintf(fmtAutheliaBuild, utils.BuildTag, utils.BuildState, utils.BuildBranch, utils.BuildCommit,
		utils.BuildNumber, runtime.GOOS, runtime.GOARCH, runtime.Compiler, utils.BuildDate, utils.BuildExtra))

	var verbose bool

	b.WriteString(fmt.Sprintf("\n"+fmtAutheliaBuildGo, info.GoVersion, info.Main.Path, info.Path))

	if verbose, err = cmd.Flags().GetBool("verbose"); err == nil && verbose {
		if len(info.Settings) != 0 {
			b.WriteString("    Settings:\n")

			for _, setting := range info.Settings {
				b.WriteString(fmt.Sprintf("        %s: %s\n", setting.Key, setting.Value))
			}
		}

		if len(info.Deps) != 0 {
			b.WriteString("    Dependencies:\n")

			for _, dep := range info.Deps {
				b.WriteString(fmt.Sprintf("        %s@%s (%s)\n", dep.Path, dep.Version, dep.Sum))
			}
		}
	}

	fmt.Print(b.String())

	return err
}
