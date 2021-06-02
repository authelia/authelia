package main

import (
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"plugin"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/configuration/schema"
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

func startServer() {
	logger := logging.Logger()
	config, errs := configuration.Read(configPathFlag)

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(1)
	}

	providers, nonFatalErrs, errs := configureProviders(config)
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

	server.StartServer(*config, providers)
}

func main() {
	logger := logging.Logger()
	rootCmd := &cobra.Command{
		Use: "authelia",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}

	rootCmd.Flags().StringVar(&configPathFlag, "config", "", "Configuration file")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of Authelia",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Authelia version %s, build %s\n", BuildTag, BuildCommit)
		},
	}

	rootCmd.AddCommand(versionCmd, commands.HashPasswordCmd,
		commands.ValidateConfigCmd, commands.CertificatesCmd,
		commands.RSACmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}

func configureProviders(config *schema.Configuration) (providers middlewares.Providers, nonFatalErrs, errs []error) {
	certPool, errs, nonFatalErrs := utils.NewX509CertPool(config.CertificatesDirectory)

	if len(errs) != 0 {
		return providers, nonFatalErrs, errs
	}

	storageProvider, err := configureStorageProvider(&config.Storage)
	if err != nil {
		return providers, nonFatalErrs, append(errs, err)
	}

	userProvider, err := configureUserProvider(config, certPool)
	if err != nil {
		return providers, nonFatalErrs, append(errs, err)
	}

	notifier, err := configureNotifier(config, certPool)
	if err != nil {
		return providers, nonFatalErrs, append(errs, err)
	}

	clock := utils.RealClock{}
	authorizer := authorization.NewAuthorizer(config.AccessControl)
	sessionProvider := session.NewProvider(config.Session, certPool)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)
	oidcProvider, err := oidc.NewOpenIDConnectProvider(config.IdentityProviders.OIDC)

	if err != nil {
		return providers, nonFatalErrs, append(errs, fmt.Errorf("Error initializing OpenID Connect Provider: %+v", err))
	}

	return middlewares.Providers{
		Authorizer:      authorizer,
		UserProvider:    userProvider,
		Regulator:       regulator,
		OpenIDConnect:   oidcProvider,
		StorageProvider: storageProvider,
		Notifier:        notifier,
		SessionProvider: sessionProvider,
	}, nonFatalErrs, errs
}

func configureStorageProvider(config *schema.StorageConfiguration) (provider storage.Provider, err error) {
	switch {
	case config.PostgreSQL != nil:
		provider = storage.NewPostgreSQLProvider(*config.PostgreSQL)
	case config.MySQL != nil:
		provider = storage.NewMySQLProvider(*config.MySQL)
	case config.Local != nil:
		provider = storage.NewSQLiteProvider(config.Local.Path)
	default:
		return provider, errors.New("Unrecognized storage provider")
	}

	return provider, nil
}

func configureUserProvider(config *schema.Configuration, certPool *x509.CertPool) (provider authentication.UserProvider, err error) {
	switch {
	case config.AuthenticationBackend.File != nil:
		provider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	case config.AuthenticationBackend.LDAP != nil:
		provider = authentication.NewLDAPUserProvider(*config.AuthenticationBackend.LDAP, certPool)
	case config.AuthenticationBackend.Plugin != nil:
		authenticationPlugin, err := plugin.Open(fmt.Sprintf("%s/%s.user_provider.authelia.com.so", config.PluginsDirectory, config.AuthenticationBackend.Plugin.Name))
		if err != nil {
			return provider, fmt.Errorf("Error opening authentication provider plugin: %+v", err)
		}

		up, err := authenticationPlugin.Lookup("UserProvider")
		if err != nil {
			return provider, fmt.Errorf("Error during authentication provider plugin lookup: %+v", err)
		}

		if p, ok := up.(authentication.UserProvider); !ok {
			return provider, errors.New("Error during authentication provider plugin setup: the plugin doesn't implement the interface (is it out of date)")
		} else {
			provider = p
		}
	default:
		return provider, errors.New("Unrecognized authentication provider")
	}

	return provider, nil
}

func configureNotifier(config *schema.Configuration, certPool *x509.CertPool) (provider notification.Notifier, err error) {
	switch {
	case config.Notifier.SMTP != nil:
		provider = notification.NewSMTPNotifier(*config.Notifier.SMTP, certPool)
	case config.Notifier.FileSystem != nil:
		provider = notification.NewFileNotifier(*config.Notifier.FileSystem)
	case config.Notifier.Plugin != nil:
		notifierPlugin, err := plugin.Open(fmt.Sprintf("%s/%s.notifier.authelia.com.so", config.PluginsDirectory, config.Notifier.Plugin.Name))
		if err != nil {
			return provider, fmt.Errorf("Error opening notifier provider plugin: %+v", err)
		}

		up, err := notifierPlugin.Lookup("Notifier")
		if err != nil {
			return provider, fmt.Errorf("Error during notifier provider plugin lookup: %+v", err)
		}

		if p, ok := up.(notification.Notifier); !ok {
			return provider, errors.New("Error during notifier provider plugin setup: the plugin doesn't implement the interface (is it out of date)")
		} else {
			provider = p
		}
	default:
		return provider, errors.New("Unrecognized notifier provider")
	}

	if !config.Notifier.DisableStartupCheck {
		_, err := provider.StartupCheck()
		if err != nil {
			return provider, fmt.Errorf("Error during notifier startup check: %s", err)
		}
	}

	return provider, nil
}
