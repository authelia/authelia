package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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
		Example: cmdAutheliaExample,
		Short:   fmt.Sprintf("authelia %s", version),
		Long:    fmt.Sprintf(fmtAutheliaLong, version),
		Version: version,
		Args:    cobra.NoArgs,
		PreRun:  newCmdWithConfigPreRun(true, true, true),
		Run:     cmdRootRun,
	}

	cmdWithConfigFlags(cmd, false, []string{})

	cmd.AddCommand(
		newBuildInfoCmd(),
		NewCertificatesCmd(),
		newCompletionCmd(),
		NewHashPasswordCmd(),
		NewRSACmd(),
		NewStorageCmd(),
		newValidateConfigCmd(),
		newAccessControlCommand(),
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

	doStartupChecks(config, &providers)

	s, listener := server.CreateServer(*config, providers)

	logger.Fatal(s.Serve(listener))
}

func doStartupChecks(config *schema.Configuration, providers *middlewares.Providers) {
	logger := logging.Logger()

	var (
		failures []string
		err      error
	)

	if err = doStartupCheck(logger, "storage", providers.StorageProvider, false); err != nil {
		logger.Errorf("Failure running the storage provider startup check: %+v", err)

		failures = append(failures, "storage")
	}

	if err = doStartupCheck(logger, "user", providers.UserProvider, false); err != nil {
		logger.Errorf("Failure running the user provider startup check: %+v", err)

		failures = append(failures, "user")
	}

	if err = doStartupCheck(logger, "notification", providers.Notifier, config.Notifier.DisableStartupCheck); err != nil {
		logger.Errorf("Failure running the notification provider startup check: %+v", err)

		failures = append(failures, "notification")
	}

	if !config.NTP.DisableStartupCheck && !providers.Authorizer.IsSecondFactorEnabled() {
		logger.Debug("The NTP startup check was skipped due to there being no configured 2FA access control rules")
	} else if err = doStartupCheck(logger, "ntp", providers.NTP, config.NTP.DisableStartupCheck); err != nil {
		logger.Errorf("Failure running the user provider startup check: %+v", err)

		if !config.NTP.DisableFailure {
			failures = append(failures, "ntp")
		}
	}

	if len(failures) != 0 {
		logger.Fatalf("The following providers had fatal failures during startup: %s", strings.Join(failures, ", "))
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
