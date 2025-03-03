package commands

import (
	"fmt"
	"github.com/authelia/authelia/v4/internal/service"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewRootCmd returns a new Root Cmd.
func NewRootCmd() (cmd *cobra.Command) {
	ctx := NewCmdCtx()

	version := utils.Version()

	cmd = &cobra.Command{
		Use:     "authelia",
		Short:   fmt.Sprintf(fmtCmdAutheliaShort, version),
		Long:    fmt.Sprintf(fmtCmdAutheliaLong, version),
		Example: cmdAutheliaExample,
		Version: version,
		Args:    cobra.NoArgs,
		PreRunE: ctx.ChainRunE(
			ctx.ConfigEnsureExistsRunE,
			ctx.HelperConfigLoadRunE,
			ctx.LogConfigure,
			ctx.LogProcessCurrentUserRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
			ctx.ConfigValidateLogRunE,
		),
		RunE: ctx.RootRunE,

		DisableAutoGenTag: true,
	}

	cmd.PersistentFlags().StringSliceP(cmdFlagNameConfig, "c", []string{"configuration.yml"}, "configuration files or directories to load, for more information run 'authelia -h authelia config'")
	cmd.PersistentFlags().StringSlice(cmdFlagNameConfigExpFilters, nil, "list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'")

	cmd.AddCommand(
		newAccessControlCommand(ctx),
		newBuildInfoCmd(ctx),
		newCryptoCmd(ctx),
		newStorageCmd(ctx),
		newConfigCmd(ctx),
		newConfigValidateLegacyCmd(ctx),

		newHelpTopic("config", "Help for the config file/directory paths", helpTopicConfig),
		newHelpTopic("filters", "help topic for the config filters", helpTopicConfigFilters),
		newHelpTopic("time-layouts", "help topic for the various time layouts", helpTopicTimeLayouts),
		newHelpTopic("hash-password", "help topic for hashing passwords", helpTopicHashPassword),
	)

	return cmd
}

func (ctx *CmdCtx) RootRunE(_ *cobra.Command, _ []string) (err error) {
	ctx.log.Infof("Authelia %s is starting", utils.Version())

	if os.Getenv("ENVIRONMENT") == "dev" {
		ctx.log.Info("===> Authelia is running in development mode. <===")
	}

	if err = logging.ConfigureLogger(ctx.config.Log, true); err != nil {
		ctx.log.Fatalf("Cannot configure logger: %v", err)
	}

	warns, errs := ctx.LoadProviders()

	if len(warns) != 0 {
		for _, err = range warns {
			ctx.log.Warn(err)
		}
	}

	if len(errs) != 0 {
		for _, err = range errs {
			ctx.log.Error(err)
		}

		ctx.log.Fatalf("Errors occurred provisioning providers.")
	}

	ctx.providers.StartupChecks(ctx)

	ctx.cconfig = nil

	ctx.log.Trace("Starting Services")

	service.RunAll(ctx)

	return nil
}
