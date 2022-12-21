package commands

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"

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

	cmd.PersistentFlags().StringSliceP(cmdFlagNameConfig, "c", []string{"configuration.yml"}, "configuration files to load")
	cmd.PersistentFlags().String(cmdFlagNameConfigDirectory, "", "path to a directory with yml/yaml files to load as part of the configuration")

	cmd.Flags().StringSlice(cmdFlagNameConfigExpFilters, nil, "Applies filters in order to the configuration file before the YAML parser. Options are 'template', 'expand-env'")

	cmd.AddCommand(
		newAccessControlCommand(ctx),
		newBuildInfoCmd(ctx),
		newCryptoCmd(ctx),
		newStorageCmd(ctx),
		newValidateConfigCmd(ctx),
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
		mainServer, metricsServer     *fasthttp.Server
		mainListener, metricsListener net.Listener
	)

	ctx.group.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				ctx.log.WithError(recoverErr(r)).Errorf("Critical error in server caught (recovered)")
			}
		}()

		if mainServer, mainListener, err = server.CreateDefaultServer(*ctx.config, ctx.providers); err != nil {
			return err
		}

		if err = mainServer.Serve(mainListener); err != nil {
			ctx.log.WithError(err).Error("Server (main) returned error")

			return err
		}

		return nil
	})

	ctx.group.Go(func() (err error) {
		if ctx.providers.Metrics == nil {
			return nil
		}

		defer func() {
			if r := recover(); r != nil {
				ctx.log.WithError(recoverErr(r)).Errorf("Critical error in metrics server caught (recovered)")
			}
		}()

		if metricsServer, metricsListener, err = server.CreateMetricsServer(ctx.config.Telemetry.Metrics); err != nil {
			return err
		}

		if err = metricsServer.Serve(metricsListener); err != nil {
			ctx.log.WithError(err).Error("Server (metrics) returned error")

			return err
		}

		return nil
	})

	if ctx.config.AuthenticationBackend.File != nil && ctx.config.AuthenticationBackend.File.Watch {
		provider := ctx.providers.UserProvider.(*authentication.FileUserProvider)
		if watcher, err := runServiceFileWatcher(ctx, ctx.config.AuthenticationBackend.File.Path, provider); err != nil {
			ctx.log.WithError(err).Errorf("Error opening file watcher")
		} else {
			defer watcher.Close()
		}
	}

	select {
	case s := <-quit:
		switch s {
		case syscall.SIGINT:
			ctx.log.Debugf("Shutdown started due to SIGINT")
		case syscall.SIGQUIT:
			ctx.log.Debugf("Shutdown started due to SIGQUIT")
		}
	case <-ctx.Done():
		ctx.log.Debugf("Shutdown started due to context completion")
	}

	ctx.cancel()

	ctx.log.Infof("Shutting down")

	var err error

	if mainServer != nil {
		if err = mainServer.Shutdown(); err != nil {
			ctx.log.WithError(err).Errorf("Error occurred shutting down the server")
		}
	}

	if metricsServer != nil {
		if err = metricsServer.Shutdown(); err != nil {
			ctx.log.WithError(err).Errorf("Error occurred shutting down the metrics server")
		}
	}

	if err = ctx.providers.StorageProvider.Close(); err != nil {
		ctx.log.WithError(err).Errorf("Error occurred closing the database connection")
	}

	if err = ctx.group.Wait(); err != nil {
		ctx.log.WithError(err).Errorf("Error occurred waiting for shutdown")
	}
}

type ReloadFilter func(path string) (skipped bool)

type ProviderReload interface {
	Reload() (reloaded bool, err error)
}

func runServiceFileWatcher(ctx *CmdCtx, path string, reload ProviderReload) (watcher *fsnotify.Watcher, err error) {
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}

	failed := make(chan struct{})

	var directory, filename string

	if path != "" {
		directory, filename = filepath.Dir(path), filepath.Base(path)
	}

	ctx.group.Go(func() error {
		for {
			select {
			case <-failed:
				return nil
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}

				if filename != filepath.Base(event.Name) {
					ctx.log.WithField("file", event.Name).WithField("op", event.Op).Tracef("File modification detected to irrelevant file")
					break
				}

				switch {
				case event.Op&fsnotify.Write == fsnotify.Write, event.Op&fsnotify.Create == fsnotify.Create:
					ctx.log.WithField("file", event.Name).WithField("op", event.Op).Debug("File modification detected")

					switch reloaded, err := reload.Reload(); {
					case err != nil:
						ctx.log.WithField("file", event.Name).WithField("op", event.Op).WithError(err).Error("Error occurred reloading file")
					case reloaded:
						ctx.log.WithField("file", event.Name).Info("Reloaded file successfully")
					default:
						ctx.log.WithField("file", event.Name).Debug("Reload of file was triggered but it was skipped")
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					ctx.log.WithField("file", event.Name).WithField("op", event.Op).Debug("Remove of file was detected")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				ctx.log.WithError(err).Errorf("Error while watching files")
			}
		}
	})

	if err := watcher.Add(directory); err != nil {
		failed <- struct{}{}

		return nil, err
	}

	ctx.log.WithField("directory", directory).WithField("file", filename).Debug("Directory is being watched for changes to the file")

	return watcher, nil
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
