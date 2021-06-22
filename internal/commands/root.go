package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
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

// NewRootCmd returns a new Root Cmd.
func NewRootCmd() (cmd *cobra.Command) {
	version := utils.Version()

	cmd = &cobra.Command{
		Use:     "authelia",
		Example: cmdAutheliaExample,
		Short:   fmt.Sprintf("authelia %s", version),
		Long:    fmt.Sprintf(fmtAutheliaLong, version),
		Version: version,
		PreRun:  cmdWithConfigPreRun,
		Run:     cmdRootRun,
	}

	cmdWithConfigFlags(cmd)

	cmd.AddCommand(
		newBuildCmd(),
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

	config := configuration.GetProvider().Configuration

	logger.Infof("Authelia %s is starting", utils.Version())

	if os.Getenv("ENVIRONMENT") == "dev" {
		logger.Info("===> Authelia is running in development mode. <===")
	}

	if err := logging.InitializeLogger(config.Log, true); err != nil {
		logger.Fatalf("Cannot initialize logger: %v", err)
	}

	providers, nonFatalErrs, errs := getProviders(config)
	if len(nonFatalErrs) != 0 {
		for _, err := range nonFatalErrs {
			logger.Warn(err)
		}
	}

	if len(errs) != 0 {
		for _, err := range nonFatalErrs {
			logger.Error(err)
		}

		logger.Fatalf("Errors occurred provisioning providers.")
	}

	server.StartServer(*config, providers)
}

func getProviders(config *schema.Configuration) (providers middlewares.Providers, nonFatalErrs []error, errs []error) {
	autheliaCertPool, errs, nonFatalErrs := utils.NewX509CertPool(config.CertificatesDirectory)
	if len(errs) != 0 || len(nonFatalErrs) != 0 {
		return providers, nonFatalErrs, errs
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
		errs = append(errs, fmt.Errorf("unrecognized storage provider"))
	}

	var userProvider authentication.UserProvider

	switch {
	case config.AuthenticationBackend.File != nil:
		userProvider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	case config.AuthenticationBackend.LDAP != nil:
		userProvider = authentication.NewLDAPUserProvider(*config.AuthenticationBackend.LDAP, autheliaCertPool)
	default:
		errs = append(errs, fmt.Errorf("unrecognized user provider"))
	}

	var notifier notification.Notifier

	switch {
	case config.Notifier.SMTP != nil:
		notifier = notification.NewSMTPNotifier(*config.Notifier.SMTP, autheliaCertPool)
	case config.Notifier.FileSystem != nil:
		notifier = notification.NewFileNotifier(*config.Notifier.FileSystem)
	default:
		errs = append(errs, fmt.Errorf("unrecognized notifier provider"))
	}

	clock := utils.RealClock{}
	authorizer := authorization.NewAuthorizer(config)
	sessionProvider := session.NewProvider(config.Session, autheliaCertPool)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)

	oidcProvider, err := oidc.NewOpenIDConnectProvider(config.IdentityProviders.OIDC)
	if err != nil {
		errs = append(errs, err)
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

// cmdWithConfigFlags is used for commands which require access to the configuration to add the flag to the command.
func cmdWithConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("config", "c", []string{}, "Configuration files")
}

// cmdWithConfigPreRun is used for commands which require access to the configuration to load the configuration in the PreRun.
func cmdWithConfigPreRun(cmd *cobra.Command, _ []string) {
	if cmd.Name() == "help" {
		return
	}

	logger := logging.Logger()

	configs, err := cmd.Root().Flags().GetStringSlice("config")
	if err != nil {
		logger.Fatalf("Error reading flags: %v", err)
	}

	provider := configuration.GetProvider()

	err = provider.LoadPaths(configs)
	if err != nil {
		logger.Fatalf("Error loading file configuration: %v", err)
	}

	err = provider.LoadEnvironment()
	if err != nil {
		logger.Fatalf("Error loading environment configuration: %v", err)
	}

	err = provider.LoadSecrets()
	if err != nil {
		for _, err := range provider.Errors() {
			logger.Errorf("\t%+v", err)
		}

		logger.Fatalf("Errors loading secrets configuration: %v", err)
	}

	err = provider.UnmarshalToStruct()
	if err != nil {
		logger.Fatalf("Error unmarshalling configuration: %v", err)
	}

	provider.ValidateConfiguration()

	warns := provider.Warnings()
	if len(warns) != 0 {
		logger.Warnf("Warnings occurred while validating configuration:")

		for _, warn := range warns {
			logger.Warnf("\t%v", warn)
		}
	}

	errs := provider.Errors()
	if len(errs) != 0 {
		logger.Errorf("Errors occurred while validating configuration:")

		for _, err := range errs {
			logger.Errorf("\t%v", err)
		}

		logger.Fatalf("Exiting due to configuration validation errors above.")
	}
}
