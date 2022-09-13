package commands

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/server"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewRootCmd returns a new Root Cmd.
func NewRootCmd() (cmd *cobra.Command) {
	version := utils.Version()

	cmd = &cobra.Command{
		Use:     "authelia",
		Short:   fmt.Sprintf(fmtCmdAutheliaShort, version),
		Long:    fmt.Sprintf(fmtCmdAutheliaLong, version),
		Example: cmdAutheliaExample,
		Version: version,
		Args:    cobra.NoArgs,
		PreRun:  newCmdWithConfigPreRun(true, true, true),
		Run:     cmdRootRun,

		DisableAutoGenTag: true,
	}

	cmdWithConfigFlags(cmd, false, []string{})

	cmd.AddCommand(
		newAccessControlCommand(),
		newBuildInfoCmd(),
		newCryptoCmd(),
		newHashPasswordCmd(),
		newStorageCmd(),
		newValidateConfigCmd(),
	)

	return cmd
}

func cmdRootRun(_ *cobra.Command, _ []string) {
	logger := logging.Logger()

	logger.Infof("Authelia %s is starting", utils.Version())

	if os.Getenv("ENVIRONMENT") == "dev" {
		logger.Info("===> Authelia is running in development mode. <===")
	}

	if err := logging.InitializeLogger(config.Log, true); err != nil {
		logger.Fatalf("Cannot initialize logger: %v", err)
	}

	providers, warnings, errors := getProviders()
	if len(warnings) != 0 {
		for _, err := range warnings {
			logger.Warn(err)
		}
	}

	if len(errors) != 0 {
		for _, err := range errors {
			logger.Error(err)
		}

		logger.Fatalf("Errors occurred provisioning providers.")
	}

	doStartupChecks(config, &providers, logger)

	runServers(config, providers, logger)
}

//nolint:gocyclo // Complexity is required in this function.
func runServers(config *schema.Configuration, providers middlewares.Providers, log *logrus.Logger) {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(quit)

	g, ctx := errgroup.WithContext(ctx)

	var (
		mainServer, metricsServer     *fasthttp.Server
		mainListener, metricsListener net.Listener
	)

	g.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.WithError(recoverErr(r)).Errorf("Critical error in server caught (recovered)")
			}
		}()

		if mainServer, mainListener, err = server.CreateDefaultServer(*config, providers); err != nil {
			return err
		}

		if err = mainServer.Serve(mainListener); err != nil {
			log.WithError(err).Error("Server (main) returned error")

			return err
		}

		return nil
	})

	g.Go(func() (err error) {
		if providers.Metrics == nil {
			return nil
		}

		defer func() {
			if r := recover(); r != nil {
				log.WithError(recoverErr(r)).Errorf("Critical error in metrics server caught (recovered)")
			}
		}()

		if metricsServer, metricsListener, err = server.CreateMetricsServer(config.Telemetry.Metrics); err != nil {
			return err
		}

		if err = metricsServer.Serve(metricsListener); err != nil {
			log.WithError(err).Error("Server (metrics) returned error")

			return err
		}

		return nil
	})

	select {
	case s := <-quit:
		switch s {
		case syscall.SIGINT:
			log.Debugf("Shutdown started due to SIGINT")
		case syscall.SIGQUIT:
			log.Debugf("Shutdown started due to SIGQUIT")
		}
	case <-ctx.Done():
		log.Debugf("Shutdown started due to context completion")
	}

	cancel()

	log.Infof("Shutting down")

	var err error

	if mainServer != nil {
		if err = mainServer.Shutdown(); err != nil {
			log.WithError(err).Errorf("Error occurred shutting down the server")
		}
	}

	if metricsServer != nil {
		if err = metricsServer.Shutdown(); err != nil {
			log.WithError(err).Errorf("Error occurred shutting down the metrics server")
		}
	}

	if err = g.Wait(); err != nil {
		log.WithError(err).Errorf("Error occurred waiting for shutdown")
	}
}

func doStartupChecks(config *schema.Configuration, providers *middlewares.Providers, log *logrus.Logger) {
	var (
		failures []string
		err      error
	)

	if err = doStartupCheck(log, "storage", providers.StorageProvider, false); err != nil {
		log.Errorf("Failure running the storage provider startup check: %+v", err)

		failures = append(failures, "storage")
	}

	if err = doStartupCheck(log, "user", providers.UserProvider, false); err != nil {
		log.Errorf("Failure running the user provider startup check: %+v", err)

		failures = append(failures, "user")
	}

	if err = doStartupCheck(log, "notification", providers.Notifier, config.Notifier.DisableStartupCheck); err != nil {
		log.Errorf("Failure running the notification provider startup check: %+v", err)

		failures = append(failures, "notification")
	}

	if !config.NTP.DisableStartupCheck && !providers.Authorizer.IsSecondFactorEnabled() {
		log.Debug("The NTP startup check was skipped due to there being no configured 2FA access control rules")
	} else if err = doStartupCheck(log, "ntp", providers.NTP, config.NTP.DisableStartupCheck); err != nil {
		log.Errorf("Failure running the ntp provider startup check: %+v", err)

		if !config.NTP.DisableFailure {
			failures = append(failures, "ntp")
		}
	}

	if len(failures) != 0 {
		log.Fatalf("The following providers had fatal failures during startup: %s", strings.Join(failures, ", "))
	}
}

func doStartupCheck(logger *logrus.Logger, name string, provider model.StartupCheck, disabled bool) error {
	if disabled {
		logger.Debugf("%s provider: startup check skipped as it is disabled", name)
		return nil
	}

	if provider == nil {
		return fmt.Errorf("unrecognized provider or it is not configured properly")
	}

	return provider.StartupCheck()
}
