package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/kubernetes/v1/clientset"
	"github.com/authelia/authelia/internal/kubernetes/v1/types"
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

//nolint:gocyclo // TODO: Consider refactoring/simplifying, time permitting.
func startServer() {
	logger := logging.Logger()
	config, errs := configuration.Read(configPathFlag)

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(1)
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

	if err := logging.InitializeLogger(config.LogFormat, config.LogFilePath); err != nil {
		logger.Fatalf("Cannot initialize logger: %v", err)
	}

	switch config.LogLevel {
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
	authorizer := authorization.NewAuthorizer(config.AccessControl)
	sessionProvider := session.NewProvider(config.Session, autheliaCertPool)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)
	oidcProvider, err := oidc.NewOpenIDConnectProvider(config.IdentityProviders.OIDC)

	if err != nil {
		logger.Fatalf("Error initializing OpenID Connect Provider: %+v", err)
	}

	if config.Kubernetes.IsEnabled() {
		logger.Debug("Creating Kubernetes client")
		var kubernetesConfig *rest.Config
		if config.Kubernetes.UseFlags() {
			kubernetesConfig, err = clientcmd.BuildConfigFromFlags(config.Kubernetes.MasterURL, config.Kubernetes.ConfigFilePath)
		} else {
			kubernetesConfig, err = rest.InClusterConfig()
		}
		if err != nil {
			logger.Fatalf("Unable to configure Kubernetes client: %+v", err)
		}

		types.AddToScheme(scheme.Scheme)

		kubernetesClient, err := clientset.NewClient(kubernetesConfig)
		if err != nil {
			logger.Fatalf("Unable to create Kubernetes client: %+v", err)
		}

		if config.Kubernetes.TrustAccessControlRules {
			logger.Debug("Enabling Kubernetes AccessControlRule watcher")
			informer := kubernetesClient.AccessControlRules().Namespace(config.Kubernetes.Namespace).CreateInformer()
			informer.AddFunc = func(rule *types.AccessControlRule) {
				logger.Println("=== Rule Added ===")
				logger.Println(rule.Spec.Domains)
				logger.Println(rule.Spec.Policy)
				logger.Println(rule.Spec.Subjects)
				logger.Println(rule.ResourceVersion)
				logger.Println("==================")
			}
			informer.UpdateFunc = func(oldRule *types.AccessControlRule, newRule *types.AccessControlRule) {
				logger.Println("=== Rule Updated ===")
				logger.Println("Old:")
				logger.Println(oldRule.Spec.Domains)
				logger.Println(oldRule.Spec.Policy)
				logger.Println(oldRule.Spec.Subjects)
				logger.Println(oldRule.ResourceVersion)
				logger.Println("New:")
				logger.Println(newRule.Spec.Domains)
				logger.Println(newRule.Spec.Policy)
				logger.Println(newRule.Spec.Subjects)
				logger.Println(newRule.ResourceVersion)
				logger.Println("====================")
			}
			informer.DeleteFunc = func(rule *types.AccessControlRule) {
				logger.Println("=== Rule Deleted ===")
				logger.Println(rule.Spec.Domains)
				logger.Println(rule.Spec.Policy)
				logger.Println(rule.Spec.Subjects)
				logger.Println(rule.ResourceVersion)
				logger.Println("====================")
			}

			informer.Start()

			logger.Debug("Waiting for initial Kubernetes sync")
			err = informer.WaitForSync()
			if err != nil {
				logger.Fatalf("Unable to perform initial Kubernetes synchronization")
			}
		}
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
