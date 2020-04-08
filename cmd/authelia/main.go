package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/notification"
	"github.com/authelia/authelia/internal/regulation"
	"github.com/authelia/authelia/internal/server"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/storage"
	"github.com/authelia/authelia/internal/utils"
)

var configPathFlag string

func startServer() {
	if configPathFlag == "" {
		log.Fatal(errors.New("No config file path provided"))
	}

	config, errs := configuration.Read(configPathFlag)

	if len(errs) > 0 {
		for _, err := range errs {
			logging.Logger().Error(err)
		}
		panic(errors.New("Some errors have been reported"))
	}

	if err := logging.InitializeLogger(config.LogFilePath); err != nil {
		log.Fatalf("Cannot initialize logger: %v", err)
	}

	switch config.LogLevel {
	case "info":
		logging.Logger().Info("Logging severity set to info")
		logging.SetLevel(logrus.InfoLevel)
		break
	case "debug":
		logging.Logger().Info("Logging severity set to debug")
		logging.SetLevel(logrus.DebugLevel)
	case "trace":
		logging.Logger().Info("Logging severity set to trace")
		logging.SetLevel(logrus.TraceLevel)
	}

	if os.Getenv("ENVIRONMENT") == "dev" {
		logging.Logger().Info("===> Authelia is running in development mode. <===")
	}

	var userProvider authentication.UserProvider

	if config.AuthenticationBackend.File != nil {
		userProvider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	} else if config.AuthenticationBackend.Ldap != nil {
		userProvider = authentication.NewLDAPUserProvider(*config.AuthenticationBackend.Ldap)
	} else {
		log.Fatalf("Unrecognized authentication backend")
	}

	var storageProvider storage.Provider
	if config.Storage.PostgreSQL != nil {
		storageProvider = storage.NewPostgreSQLProvider(*config.Storage.PostgreSQL)
	} else if config.Storage.MySQL != nil {
		storageProvider = storage.NewMySQLProvider(*config.Storage.MySQL)
	} else if config.Storage.Local != nil {
		storageProvider = storage.NewSQLiteProvider(config.Storage.Local.Path)
	} else {
		log.Fatalf("Unrecognized storage backend")
	}

	var notifier notification.Notifier
	if config.Notifier.SMTP != nil {
		notifier = notification.NewSMTPNotifier(*config.Notifier.SMTP)
	} else if config.Notifier.FileSystem != nil {
		notifier = notification.NewFileNotifier(*config.Notifier.FileSystem)
	} else {
		log.Fatalf("Unrecognized notifier")
	}

	clock := utils.RealClock{}
	authorizer := authorization.NewAuthorizer(config.AccessControl)
	sessionProvider := session.NewProvider(config.Session)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)

	providers := middlewares.Providers{
		Authorizer:      authorizer,
		UserProvider:    userProvider,
		Regulator:       regulator,
		StorageProvider: storageProvider,
		Notifier:        notifier,
		SessionProvider: sessionProvider,
	}
	server.StartServer(*config, providers)
}

func main() {
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
			fmt.Printf("Authelia version %s, build %s", BuildTag, BuildCommit)
		},
	}

	rootCmd.AddCommand(versionCmd, commands.MigrateCmd, commands.HashPasswordCmd)
	rootCmd.AddCommand(commands.CertificatesCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
