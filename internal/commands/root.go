package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/notification"
	"github.com/authelia/authelia/v4/internal/ntp"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/server"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
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

	cmdWithConfigFlags(cmd)

	cmd.AddCommand(
		newBuildInfoCmd(),
		NewCertificatesCmd(),
		newCompletionCmd(),
		NewHashPasswordCmd(),
		NewRSACmd(),
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

	providers, warnings, errors := getProviders(config)
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

	server.Start(*config, providers)
}

//nolint:gocyclo // TODO: Consider refactoring time permitting.
func getProviders(config *schema.Configuration) (providers middlewares.Providers, warnings []error, errors []error) {
	logger := logging.Logger()

	autheliaCertPool, warnings, errors := utils.NewX509CertPool(config.CertificatesDirectory)
	if len(warnings) != 0 || len(errors) != 0 {
		return providers, warnings, errors
	}

	var storageProvider storage.Provider

	switch {
	case config.Storage.PostgreSQL != nil:
		storageProvider = storage.NewPostgreSQLProvider(*config.Storage.PostgreSQL)
	case config.Storage.MySQL != nil:
		storageProvider = storage.NewMySQLProvider(*config.Storage.MySQL)
	case config.Storage.Local != nil:
		storageProvider = storage.NewSQLiteProvider(config.Storage.Local.Path)
	default:
		errors = append(errors, fmt.Errorf("unrecognized storage provider"))
	}

	var (
		userProvider authentication.UserProvider
		err          error
	)

	switch {
	case config.AuthenticationBackend.File != nil:
		userProvider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	case config.AuthenticationBackend.LDAP != nil:
		userProvider, err = authentication.NewLDAPUserProvider(config.AuthenticationBackend, autheliaCertPool)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to check LDAP authentication backend: %w", err))
		}
	default:
		errors = append(errors, fmt.Errorf("unrecognized user provider"))
	}

	var notifier notification.Notifier

	switch {
	case config.Notifier.SMTP != nil:
		notifier = notification.NewSMTPNotifier(config.Notifier.SMTP, autheliaCertPool)
	case config.Notifier.FileSystem != nil:
		notifier = notification.NewFileNotifier(*config.Notifier.FileSystem)
	default:
		errors = append(errors, fmt.Errorf("unrecognized notifier provider"))
	}

	if notifier != nil && !config.Notifier.DisableStartupCheck {
		if _, err := notifier.StartupCheck(); err != nil {
			errors = append(errors, fmt.Errorf("failed to check notification provider: %w", err))
		}
	}

	var ntpProvider *ntp.Provider
	if config.NTP != nil {
		ntpProvider = ntp.NewProvider(config.NTP)
	}

	var failed bool
	if !config.NTP.DisableStartupCheck {
		failed, err = ntpProvider.StartupCheck()

		if err != nil {
			logger.Errorf("Failed to check NTP host-server synchronization: %v", err)
		}

		if failed {
			logger.Fatalf("Your Host is not synchronized with an NTP server")
		} else {
			logger.Info("*==NTP synchronization check terminated with success==*")
		}
	}

	clock := utils.RealClock{}
	authorizer := authorization.NewAuthorizer(config)
	sessionProvider := session.NewProvider(config.Session, autheliaCertPool)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)

	oidcProvider, err := oidc.NewOpenIDConnectProvider(config.IdentityProviders.OIDC)
	if err != nil {
		errors = append(errors, err)
	}

	return middlewares.Providers{
		Authorizer:      authorizer,
		UserProvider:    userProvider,
		Regulator:       regulator,
		OpenIDConnect:   oidcProvider,
		StorageProvider: storageProvider,
		Ntp:             ntpProvider,
		Notifier:        notifier,
		SessionProvider: sessionProvider,
	}, warnings, errors
}
