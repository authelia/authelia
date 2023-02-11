package commands

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/server"
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
			ctx.ConfigLoadRunE,
			ctx.ConfigValidateKeysRunE,
			ctx.ConfigValidateRunE,
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
		newValidateConfigCmd(ctx),

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

	if err = logging.InitializeLogger(ctx.config.Log, true); err != nil {
		ctx.log.Fatalf("Cannot initialize logger: %v", err)
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

	runServices(ctx)

	return nil
}

//nolint:gocyclo // Complexity is required in this function.
func runServices(ctx *CmdCtx) {
	defer ctx.cancel()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(quit)

	var (
		services []Service

		err error
	)

	switch server, listener, err := server.CreateDefaultServer(ctx.config, ctx.providers); {
	case err != nil:
		ctx.log.WithError(err).Fatal("Create Server Service (main) returned error")
	case server != nil && listener != nil:
		svc := NewServerService("main", server, listener, ctx.log)

		ctx.group.Go(svc.Run)

		services = append(services, svc)
	default:
		ctx.log.Fatal("Create Server Service (main) failed")
	}

	switch server, listener, err := server.CreateMetricsServer(ctx.config, ctx.providers); {
	case err != nil:
		ctx.log.WithError(err).Fatal("Create Server Service (metrics) returned error")
	case server != nil && listener != nil:
		svc := NewServerService("metrics", server, listener, ctx.log)

		ctx.group.Go(svc.Run)

		services = append(services, svc)
	default:
		ctx.log.Debug("Create Server Service (metrics) skipped")
	}

	if ctx.config.AuthenticationBackend.File != nil && ctx.config.AuthenticationBackend.File.Watch {
		provider := ctx.providers.UserProvider.(*authentication.FileUserProvider)

		var svc *FileWatcherService

		if svc, err = NewFileWatcherService("users", ctx.config.AuthenticationBackend.File.Path, provider, ctx.log); err != nil {
			ctx.log.WithError(err).Fatal("Create Watcher Service (users) returned error")
		}

		ctx.group.Go(svc.Run)

		services = append(services, svc)
	}

	select {
	case s := <-quit:
		switch s {
		case syscall.SIGINT:
			ctx.log.WithField("signal", "SIGINT").Debugf("Shutdown started due to signal")
		case syscall.SIGTERM:
			ctx.log.WithField("signal", "SIGTERM").Debugf("Shutdown started due to signal")
		}
	case <-ctx.Done():
		ctx.log.Debugf("Shutdown started due to context completion")
	}

	ctx.cancel()

	ctx.log.Infof("Shutting down")

	wgShutdown := &sync.WaitGroup{}

	for _, service := range services {
		go func() {
			service.Shutdown()

			wgShutdown.Done()
		}()

		wgShutdown.Add(1)
	}

	wgShutdown.Wait()

	if err = ctx.providers.StorageProvider.Close(); err != nil {
		ctx.log.WithError(err).Error("Error occurred closing database connections")
	}

	if err = ctx.group.Wait(); err != nil {
		ctx.log.WithError(err).Errorf("Error occurred waiting for shutdown")
	}
}

func doStartupChecks(ctx *CmdCtx) {
	var (
		failures []string
		err      error
	)

	if err = doStartupCheck(ctx, "storage", ctx.providers.StorageProvider, false); err != nil {
		ctx.log.Errorf("Failure running the storage provider startup check: %+v", err)

		failures = append(failures, "storage")
	}

	if err = doStartupCheck(ctx, "user", ctx.providers.UserProvider, false); err != nil {
		ctx.log.Errorf("Failure running the user provider startup check: %+v", err)

		failures = append(failures, "user")
	}

	if err = doStartupCheck(ctx, "notification", ctx.providers.Notifier, ctx.config.Notifier.DisableStartupCheck); err != nil {
		ctx.log.Errorf("Failure running the notification provider startup check: %+v", err)

		failures = append(failures, "notification")
	}

	if !ctx.config.NTP.DisableStartupCheck && !ctx.providers.Authorizer.IsSecondFactorEnabled() {
		ctx.log.Debug("The NTP startup check was skipped due to there being no configured 2FA access control rules")
	} else if err = doStartupCheck(ctx, "ntp", ctx.providers.NTP, ctx.config.NTP.DisableStartupCheck); err != nil {
		ctx.log.Errorf("Failure running the ntp provider startup check: %+v", err)

		if !ctx.config.NTP.DisableFailure {
			failures = append(failures, "ntp")
		}
	}

	if len(failures) != 0 {
		ctx.log.Fatalf("The following providers had fatal failures during startup: %s", strings.Join(failures, ", "))
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
