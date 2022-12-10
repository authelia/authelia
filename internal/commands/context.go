package commands

import (
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/notification"
	"github.com/authelia/authelia/v4/internal/ntp"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/totp"
	"github.com/authelia/authelia/v4/internal/utils"
)

func NewCommandContext() *CmdCtx {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)

	group, ctx := errgroup.WithContext(ctx)

	return &CmdCtx{
		Context: ctx,
		cancel:  cancel,
		group:   group,
		log:     logging.Logger(),
		config:  &schema.Configuration{},
	}
}

type CmdCtx struct {
	context.Context

	cancel context.CancelFunc
	group  *errgroup.Group

	log *logrus.Logger

	config    *schema.Configuration
	providers middlewares.Providers
	trusted   *x509.CertPool

	cconfig *CommandContextConfig
}

func NewCommandContextConfig() *CommandContextConfig {
	return &CommandContextConfig{
		validator: schema.NewStructValidator(),
	}
}

type CommandContextConfig struct {
	defaults  configuration.Source
	sources   []configuration.Source
	keys      []string
	validator *schema.StructValidator
}

type CobraRunECmd func(cmd *cobra.Command, args []string) (err error)

func (ctx *CmdCtx) ProvidersLoadStoragePreRunE(cmd *cobra.Command, args []string) (err error) {
	switch warns, errs := ctx.LoadTrustedCertificates(); {
	case len(errs) != 0:
		err = fmt.Errorf("had the following errors loading the trusted certificates")

		for _, e := range errs {
			err = fmt.Errorf("%+v: %w", err, e)
		}

		return err
	case len(warns) != 0:
		err = fmt.Errorf("had the following warnings loading the trusted certificates")

		for _, e := range errs {
			err = fmt.Errorf("%+v: %w", err, e)
		}

		return err
	default:
		ctx.providers.StorageProvider = getStorageProvider(ctx)

		return nil
	}
}

func (ctx *CmdCtx) CheckSchemaVersion() (err error) {
	if ctx.providers.StorageProvider == nil {
		return fmt.Errorf("storage not loaded")
	}

	var version, latest int

	if version, err = ctx.providers.StorageProvider.SchemaVersion(ctx); err != nil {
		return err
	}

	if latest, err = ctx.providers.StorageProvider.SchemaLatestVersion(); err != nil {
		return err
	}

	switch {
	case version > latest:
		return fmt.Errorf("%w: version %d is not compatible with this version of the binary as the latest compatible version is %d", errStorageSchemaIncompatible, version, latest)
	case version < latest:
		return fmt.Errorf("%w: version %d is outdated please migrate to version %d in order to use this command or use an older binary", errStorageSchemaOutdated, version, latest)
	default:
		return nil
	}
}

func (ctx *CmdCtx) LoadTrustedCertificates() (warns, errs []error) {
	ctx.trusted, warns, errs = utils.NewX509CertPool(ctx.config.CertificatesDirectory)

	return warns, errs
}

func (ctx *CmdCtx) ProvidersLoad() (warns, errs []error) {
	// TODO: Adjust this so the CertPool can be used like a provider.
	if warns, errs = ctx.LoadTrustedCertificates(); len(warns) != 0 || len(errs) != 0 {
		return warns, errs
	}

	clock := utils.RealClock{}

	providers := middlewares.Providers{
		Authorizer:      authorization.NewAuthorizer(ctx.config),
		NTP:             ntp.NewProvider(&ctx.config.NTP),
		PasswordPolicy:  middlewares.NewPasswordPolicyProvider(ctx.config.PasswordPolicy),
		SessionProvider: session.NewProvider(ctx.config.Session, ctx.trusted),
		StorageProvider: getStorageProvider(ctx),
		TOTP:            totp.NewTimeBasedProvider(ctx.config.TOTP),
	}

	providers.Regulator = regulation.NewRegulator(ctx.config.Regulation, providers.StorageProvider, clock)

	var err error

	switch {
	case ctx.config.AuthenticationBackend.File != nil:
		providers.UserProvider = authentication.NewFileUserProvider(ctx.config.AuthenticationBackend.File)
	case ctx.config.AuthenticationBackend.LDAP != nil:
		providers.UserProvider = authentication.NewLDAPUserProvider(ctx.config.AuthenticationBackend, ctx.trusted)
	}

	if providers.Templates, err = templates.New(templates.Config{EmailTemplatesPath: ctx.config.Notifier.TemplatePath}); err != nil {
		errs = append(errs, err)
	}

	switch {
	case ctx.config.Notifier.SMTP != nil:
		providers.Notifier = notification.NewSMTPNotifier(ctx.config.Notifier.SMTP, ctx.trusted, providers.Templates)
	case ctx.config.Notifier.FileSystem != nil:
		providers.Notifier = notification.NewFileNotifier(*ctx.config.Notifier.FileSystem)
	}

	if providers.OpenIDConnect, err = oidc.NewOpenIDConnectProvider(ctx.config.IdentityProviders.OIDC, providers.StorageProvider); err != nil {
		errs = append(errs, err)
	}

	if ctx.config.Telemetry.Metrics.Enabled {
		providers.Metrics = metrics.NewPrometheus()
	}

	ctx.providers = providers

	return warns, errs
}

func (ctx *CmdCtx) ChainPreRunE(cmdRunEs ...CobraRunECmd) CobraRunECmd {
	return func(cmd *cobra.Command, args []string) (err error) {
		for _, cmdRunE := range cmdRunEs {
			if err = cmdRunE(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}

func (ctx *CmdCtx) ConfigSetFlagsMapRunE(flags *pflag.FlagSet, flagsMap map[string]string, includeInvalidKeys, includeUnchangedKeys bool) (err error) {
	if ctx.cconfig == nil {
		ctx.cconfig = NewCommandContextConfig()
	}

	ctx.cconfig.sources = append(ctx.cconfig.sources, configuration.NewCommandLineSourceWithMapping(flags, flagsMap, includeInvalidKeys, includeUnchangedKeys))

	return nil
}

func (ctx *CmdCtx) ConfigSetMapDefaultsRunE(defaults map[string]any) CobraRunECmd {
	return func(cmd *cobra.Command, args []string) (err error) {
		if ctx.cconfig == nil {
			ctx.cconfig = NewCommandContextConfig()
		}

		ctx.cconfig.defaults = configuration.NewMapSource(defaults)

		return nil
	}
}

func (ctx *CmdCtx) ConfigValidateKeysPreRunE(_ *cobra.Command, _ []string) (err error) {
	if ctx.cconfig == nil {
		return fmt.Errorf("config validate keys must be used with ConfigLoadPreRunE")
	}

	validator.ValidateKeys(ctx.cconfig.keys, configuration.DefaultEnvPrefix, ctx.cconfig.validator)

	return nil
}

func (ctx *CmdCtx) ConfigValidatePreRunE(_ *cobra.Command, _ []string) (err error) {
	validator.ValidateConfiguration(ctx.config, ctx.cconfig.validator)

	return nil
}

func (ctx *CmdCtx) ConfigValidateLogPreRunE(_ *cobra.Command, _ []string) (err error) {
	warnings := ctx.cconfig.validator.Warnings()
	if len(warnings) != 0 {
		for _, warning := range warnings {
			ctx.log.Warnf("Configuration: %+v", warning)
		}
	}

	errs := ctx.cconfig.validator.Errors()
	if len(errs) != 0 {
		for _, err = range errs {
			ctx.log.Errorf("Configuration: %+v", err)
		}

		ctx.log.Fatalf("Can't continue due to the errors loading the configuration")
	}

	return nil
}

func (ctx *CmdCtx) ConfigEnsureExistsPreRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		configs   []string
		directory string
		created   bool
	)

	if configs, _, err = loadEnvCLIStringSliceValue(cmd, "X_AUTHELIA_CONFIG", cmdFlagNameConfig); err != nil {
		return err
	}

	if directory, _, err = loadEnvCLIStringValue(cmd, "X_AUTHELIA_CONFIG_DIRECTORY", cmdFlagNameConfigDirectory); err != nil {
		return err
	}

	if len(configs) != 1 || directory != "" {
		return nil
	}

	if created, err = configuration.EnsureConfigurationExists(configs[0]); err != nil {
		ctx.log.Fatal(err)
	}

	if created {
		ctx.log.Warnf("Configuration did not exist so a default one has been generated at %s, you will need to configure this", configs[0])
		os.Exit(0)
	}

	return nil
}

func (ctx *CmdCtx) ConfigLoadPreRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		configs, filterNames []string

		filters []configuration.FileFilter

		explicit bool

		directory   string
		explicitDir bool
	)

	if configs, explicit, err = loadEnvCLIStringSliceValue(cmd, "X_AUTHELIA_CONFIG", cmdFlagNameConfig); err != nil {
		return err
	}

	if directory, explicitDir, err = loadEnvCLIStringValue(cmd, "X_AUTHELIA_CONFIG_DIRECTORY", cmdFlagNameConfigDirectory); err != nil {
		return err
	}

	if filterNames, err = cmd.Flags().GetStringSlice(cmdFlagNameConfigExpFilters); err != nil {
		ctx.log.Fatalf("Error reading flags: %v", err)
	}

	if filters, err = configuration.NewFileFilters(filterNames); err != nil {
		ctx.log.Fatalf("Error occurred loading configuration: flag '--%s' is invalid: %v", cmdFlagNameConfigExpFilters, err)
	}

	if !explicit {
		var (
			final   []string
			changed bool
		)

		for _, c := range configs {
			if _, err = os.Stat(directory); err != nil && os.IsNotExist(err) {
				changed = true

				continue
			}

			final = append(final, c)
		}

		if changed {
			configs = final
		}
	}

	if !explicitDir {
		if _, err = os.Stat(directory); err != nil && os.IsNotExist(err) {
			directory = ""
		}
	}

	if ctx.cconfig == nil {
		ctx.cconfig = NewCommandContextConfig()
	}

	if ctx.cconfig.keys, err = configuration.LoadAdvanced(
		ctx.cconfig.validator,
		"",
		ctx.config,
		configuration.NewDefaultSourcesWithDefaults(
			configs,
			directory,
			filters,
			configuration.DefaultEnvPrefix,
			configuration.DefaultEnvDelimiter,
			ctx.cconfig.defaults,
			ctx.cconfig.sources...)...); err != nil {
		return err
	}

	return nil
}

func loadEnvCLIStringValue(cmd *cobra.Command, envKey, flagName string) (value string, explicit bool, err error) {
	env, ok := os.LookupEnv(envKey)

	switch {
	case cmd.Flags().Changed(flagName):
		value, err = cmd.Flags().GetString(flagName)

		return value, true, err
	case ok && env != "":
		return env, true, nil
	default:
		value, err = cmd.Flags().GetString(flagName)

		return value, false, err
	}
}

func loadEnvCLIStringSliceValue(cmd *cobra.Command, envKey, flagName string) (value []string, explicit bool, err error) {
	env, ok := os.LookupEnv(envKey)

	switch {
	case cmd.Flags().Changed(flagName):
		value, err = cmd.Flags().GetStringSlice(flagName)

		return value, true, err
	case ok && env != "":
		return strings.Split(env, ","), true, nil
	default:
		value, err = cmd.Flags().GetStringSlice(flagName)

		return value, false, err
	}
}
