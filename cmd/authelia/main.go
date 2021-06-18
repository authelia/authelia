package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/notification"
	"github.com/authelia/authelia/internal/oidc"
	"github.com/authelia/authelia/internal/regulation"
	"github.com/authelia/authelia/internal/server"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/storage"
	"github.com/authelia/authelia/internal/utils"
)

var configPathFlag string

func main() {
	logger := logging.Logger()
	version := utils.Version()

	rootCmd := &cobra.Command{
		Use:               "authelia",
		Example:           cmdAutheliaExample,
		Short:             fmt.Sprintf("authelia %s", version),
		Long:              fmt.Sprintf(fmtAutheliaLong, version),
		Version:           version,
		RunE:              startServer,
		PersistentPreRunE: preRun,
	}

	rootCmd.PersistentFlags().StringSliceP("config", "c", []string{}, "Configuration files")

	// TODO: Add configuration flags here.

	rootCmd.AddCommand(buildCmd, completionCmd, commands.HashPasswordCmd,
		commands.ValidateConfigCmd, commands.CertificatesCmd,
		commands.RSACmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}

//nolint:gocyclo // TODO: Consider refactoring/simplifying, time permitting.
func startServer(cmd *cobra.Command, args []string) error {
	logger := logging.Logger()
	config, errs := configuration.Read(configPathFlag)

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		return errors.New("error occurred during config validation")
	}

	autheliaCertPool, errs, nonFatalErrs := utils.NewX509CertPool(config.CertificatesDirectory)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(2)
	}

	if len(nonFatalErrs) > 0 {
		for _, err := range nonFatalErrs {
			logger.Warn(err)
		}
	}

	if err := logging.InitializeLogger(config.Logging.Format, config.Logging.FilePath, config.Logging.KeepStdout); err != nil {
		logger.Fatalf("Cannot initialize logger: %v", err)
	}

	logger.Infof("Authelia %s is starting", utils.Version())

	switch config.Logging.Level {
	case "error":
		logger.Info("Logging severity set to error")
		logging.SetLevel(logrus.ErrorLevel)
	case "warn":
		logger.Info("Logging severity set to warn")
		logging.SetLevel(logrus.WarnLevel)
	case "info":
		logger.Info("Logging severity set to info")
		logging.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.Info("Logging severity set to debug")
		logging.SetLevel(logrus.DebugLevel)
	case "trace":
		logger.Info("Logging severity set to trace")
		logging.SetLevel(logrus.TraceLevel)
	}

	if os.Getenv("ENVIRONMENT") == "dev" {
		logger.Info("===> Authelia is running in development mode. <===")
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
		logger.Fatalf("Unrecognized storage backend")
	}

	var userProvider authentication.UserProvider

	switch {
	case config.AuthenticationBackend.File != nil:
		userProvider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	case config.AuthenticationBackend.LDAP != nil:
		userProvider = authentication.NewLDAPUserProvider(*config.AuthenticationBackend.LDAP, autheliaCertPool)
	default:
		logger.Fatalf("Unrecognized authentication backend")
	}

	var notifier notification.Notifier

	switch {
	case config.Notifier.SMTP != nil:
		notifier = notification.NewSMTPNotifier(*config.Notifier.SMTP, autheliaCertPool)
	case config.Notifier.FileSystem != nil:
		notifier = notification.NewFileNotifier(*config.Notifier.FileSystem)
	default:
		logger.Fatalf("Unrecognized notifier")
	}

	if !config.Notifier.DisableStartupCheck {
		_, err := notifier.StartupCheck()
		if err != nil {
			logger.Fatalf("Error during notifier startup check: %s", err)
		}
	}

	clock := utils.RealClock{}
	authorizer := authorization.NewAuthorizer(config)
	sessionProvider := session.NewProvider(config.Session, autheliaCertPool)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)
	oidcProvider, err := oidc.NewOpenIDConnectProvider(config.IdentityProviders.OIDC)

	if err != nil {
		logger.Fatalf("Error initializing OpenID Connect Provider: %+v", err)
	}

	providers := middlewares.Providers{
		Authorizer:      authorizer,
		UserProvider:    userProvider,
		Regulator:       regulator,
		OpenIDConnect:   oidcProvider,
		StorageProvider: storageProvider,
		Notifier:        notifier,
		SessionProvider: sessionProvider,
	}

	server.StartServer(*config, providers)

	return nil
}

func preRun(cmd *cobra.Command, args []string) (err error) {

	configs, err := cmd.PersistentFlags().GetStringSlice("config")
	if err != nil {
		return err
	}

	provider := configuration.GetProvider()

	err = provider.LoadFile(configs)
	if err != nil {
		return err
	}

	err = provider.LoadEnvironment()
	if err != nil {
		return err
	}

	// If running the root command we need to load Command Line Arguments.
	if cmd == cmd.Root() {
		err = provider.LoadCommandLineArguments(cmd.Flags())
		if err != nil {
			return err
		}
	}

	provider.Validate()

	warns := provider.StructValidator.Warnings()
	if len(warns) != 0 {
		for _, warn := range warns {
			logrus.Warnf(warn.Error())
		}
	}

	errs := provider.StructValidator.Errors()
	if len(errs) != 0 {
		s := strings.Builder{}

		s.WriteString("Errors during configuration validation:\n")
		for _, err := range errs {
			s.WriteString(fmt.Sprintf("  %s\n", err.Error()))
		}

		return errors.New(s.String())
	}

	return nil
}
