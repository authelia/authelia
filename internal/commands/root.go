package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
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

	doStartupChecks(ctx)

	ctx.cconfig = nil

	ctx.log.Trace("Starting Services")

	servicesRun(ctx)

	return nil
}

func doStartupChecks(ctx *CmdCtx) {
	var (
		failures []string
		err      error
	)

	ctx.log.WithFields(map[string]any{logFieldProvider: providerNameStorage}).Trace("Performing Startup Check")

	if err = doStartupCheck(ctx, providerNameStorage, ctx.providers.StorageProvider, false); err != nil {
		ctx.log.WithError(err).WithField(logFieldProvider, providerNameStorage).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameStorage)
	} else {
		ctx.log.WithFields(map[string]any{logFieldProvider: providerNameStorage}).Trace("Startup Check Completed Successfully")
	}

	ctx.log.WithFields(map[string]any{logFieldProvider: providerNameUser}).Trace("Performing Startup Check")

	if err = doStartupCheck(ctx, providerNameUser, ctx.providers.UserProvider, false); err != nil {
		ctx.log.WithError(err).WithField(logFieldProvider, providerNameUser).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameUser)
	} else {
		ctx.log.WithFields(map[string]any{logFieldProvider: providerNameUser}).Trace("Startup Check Completed Successfully")
	}

	ctx.log.WithFields(map[string]any{logFieldProvider: providerNameNotification}).Trace("Performing Startup Check")

	if err = doStartupCheck(ctx, providerNameNotification, ctx.providers.Notifier, ctx.config.Notifier.DisableStartupCheck); err != nil {
		ctx.log.WithError(err).WithField(logFieldProvider, providerNameNotification).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameNotification)
	} else {
		ctx.log.WithFields(map[string]any{logFieldProvider: providerNameNotification}).Trace("Startup Check Completed Successfully")
	}

	ctx.log.WithFields(map[string]any{logFieldProvider: providerNameNTP}).Trace("Performing Startup Check")

	if err = doStartupCheck(ctx, providerNameNTP, ctx.providers.NTP, ctx.config.NTP.DisableStartupCheck); err != nil {
		if !ctx.config.NTP.DisableFailure {
			ctx.log.WithError(err).WithField(logFieldProvider, providerNameNTP).Error(logMessageStartupCheckError)

			failures = append(failures, providerNameNTP)
		} else {
			ctx.log.WithError(err).WithField(logFieldProvider, providerNameNTP).Warn(logMessageStartupCheckError)
		}
	} else {
		ctx.log.WithFields(map[string]any{logFieldProvider: providerNameNTP}).Trace("Startup Check Completed Successfully")
	}

	if len(failures) != 0 {
		ctx.log.WithField("providers", failures).Fatalf("One or more providers had fatal failures performing startup checks, for more detail check the error level logs")
	}
}

func doStartupCheck(ctx *CmdCtx, name string, provider model.StartupCheck, disabled bool) error {
	if disabled {
		ctx.log.Debugf("%s provider: startup check skipped as it is disabled", name)
		return nil
	}

	if provider == nil {
		return fmt.Errorf("unrecognized provider or it is not configured properly")
	}

	return provider.StartupCheck()
}
