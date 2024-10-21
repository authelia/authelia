package commands

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/notification"
	"github.com/authelia/authelia/v4/internal/ntp"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/totp"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewCmdCtx returns a new CmdCtx.
func NewCmdCtx() *CmdCtx {
	ctx := context.Background()

	return &CmdCtx{
		Context: ctx,
		log:     logging.Logger(),
		providers: middlewares.Providers{
			Random: &random.Cryptographical{},
		},
		config: &schema.Configuration{},
	}
}

// CmdCtx is a context.Context used for the root command.
type CmdCtx struct {
	context.Context

	log *logrus.Logger

	config    *schema.Configuration
	providers middlewares.Providers
	trusted   *x509.CertPool

	cconfig *CmdCtxConfig
}

// NewCmdCtxConfig returns a new CmdCtxConfig.
func NewCmdCtxConfig() *CmdCtxConfig {
	return &CmdCtxConfig{
		validator: schema.NewStructValidator(),
	}
}

// CmdCtxConfig is the configuration for the CmdCtx.
type CmdCtxConfig struct {
	files     []string
	filters   []string
	defaults  configuration.Source
	sources   []configuration.Source
	keys      []string
	validator *schema.StructValidator
}

// CobraRunECmd describes a function that can be used as a *cobra.Command RunE, PreRunE, or PostRunE.
type CobraRunECmd func(cmd *cobra.Command, args []string) (err error)

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

// CheckSchema is a utility function which checks the schema version and encryption key.
func (ctx *CmdCtx) CheckSchema() (err error) {
	if err = ctx.CheckSchemaVersion(); err != nil {
		return err
	}

	var result storage.EncryptionValidationResult

	if result, err = ctx.providers.StorageProvider.SchemaEncryptionCheckKey(ctx, false); !result.Checked() || !result.Success() {
		if err != nil {
			return fmt.Errorf("failed to check the schema encryption key: %w", err)
		}

		return fmt.Errorf("failed to check the schema encryption key: the key is not valid for the schema")
	}

	return nil
}

// LoadTrustedCertificates loads the trusted certificates into the CmdCtx.
func (ctx *CmdCtx) LoadTrustedCertificates() (warns, errs []error) {
	ctx.trusted, warns, errs = utils.NewX509CertPool(ctx.config.CertificatesDirectory)

	return warns, errs
}

// LoadProviders loads all providers into the CmdCtx.
func (ctx *CmdCtx) LoadProviders() (warns, errs []error) {
	// TODO: Adjust this so the CertPool can be used like a provider.
	if warns, errs = ctx.LoadTrustedCertificates(); len(warns) != 0 || len(errs) != 0 {
		return warns, errs
	}

	ctx.providers.StorageProvider = getStorageProvider(ctx)

	ctx.providers.Authorizer = authorization.NewAuthorizer(ctx.config)
	ctx.providers.NTP = ntp.NewProvider(&ctx.config.NTP)
	ctx.providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(ctx.config.PasswordPolicy)
	ctx.providers.Regulator = regulation.NewRegulator(ctx.config.Regulation, ctx.providers.StorageProvider, clock.New())
	ctx.providers.SessionProvider = session.NewProvider(ctx.config.Session, ctx.trusted)
	ctx.providers.TOTP = totp.NewTimeBasedProvider(ctx.config.TOTP)

	var err error

	switch {
	case ctx.config.AuthenticationBackend.File != nil:
		ctx.providers.UserProvider = authentication.NewFileUserProvider(ctx.config.AuthenticationBackend.File)
	case ctx.config.AuthenticationBackend.LDAP != nil:
		ctx.providers.UserProvider = authentication.NewLDAPUserProvider(ctx.config.AuthenticationBackend, ctx.trusted)
	}

	if ctx.providers.Templates, err = templates.New(templates.Config{EmailTemplatesPath: ctx.config.Notifier.TemplatePath}); err != nil {
		errs = append(errs, err)
	}

	switch {
	case ctx.config.Notifier.SMTP != nil:
		ctx.providers.Notifier = notification.NewSMTPNotifier(ctx.config.Notifier.SMTP, ctx.trusted)
	case ctx.config.Notifier.FileSystem != nil:
		ctx.providers.Notifier = notification.NewFileNotifier(*ctx.config.Notifier.FileSystem)
	}

	ctx.providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(ctx.config.IdentityProviders.OIDC, ctx.providers.StorageProvider, ctx.providers.Templates)

	if ctx.config.Telemetry.Metrics.Enabled {
		ctx.providers.Metrics = metrics.NewPrometheus()
	}

	return warns, errs
}

// ChainRunE runs multiple CobraRunECmd funcs one after the other returning errors.
func (ctx *CmdCtx) ChainRunE(cmdRunEs ...CobraRunECmd) CobraRunECmd {
	return func(cmd *cobra.Command, args []string) (err error) {
		for _, cmdRunE := range cmdRunEs {
			if err = cmdRunE(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}

// HelperConfigSetFlagsMapRunE adds a command line source with flags mapping.
func (ctx *CmdCtx) HelperConfigSetFlagsMapRunE(flags *pflag.FlagSet, flagsMap map[string]string, includeInvalidKeys, includeUnchangedKeys bool) (err error) {
	if ctx.cconfig == nil {
		ctx.cconfig = NewCmdCtxConfig()
	}

	ctx.cconfig.sources = append(ctx.cconfig.sources, configuration.NewCommandLineSourceWithMapping(flags, flagsMap, includeInvalidKeys, includeUnchangedKeys))

	return nil
}

// HelperConfigSetDefaultsRunE adds a defaults configuration source.
func (ctx *CmdCtx) HelperConfigSetDefaultsRunE(defaults map[string]any) CobraRunECmd {
	return func(cmd *cobra.Command, args []string) (err error) {
		if ctx.cconfig == nil {
			ctx.cconfig = NewCmdCtxConfig()
		}

		ctx.cconfig.defaults = configuration.NewMapSource(defaults)

		return nil
	}
}

// HelperConfigValidateKeysRunE validates the configuration (keys).
func (ctx *CmdCtx) HelperConfigValidateKeysRunE(_ *cobra.Command, _ []string) (err error) {
	if ctx.cconfig == nil {
		return fmt.Errorf("HelperConfigValidateKeysRunE must be used with HelperConfigLoadRunE")
	}

	validator.ValidateKeys(ctx.cconfig.keys, configuration.GetMultiKeyMappedDeprecationKeys(), configuration.DefaultEnvPrefix, ctx.cconfig.validator)

	return nil
}

// HelperConfigValidateRunE validates the configuration (structure).
func (ctx *CmdCtx) HelperConfigValidateRunE(_ *cobra.Command, _ []string) (err error) {
	tc := &tls.Config{
		RootCAs:    ctx.trusted,
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	validator.ValidateConfiguration(ctx.config, ctx.cconfig.validator, validator.WithTLSConfig(tc))

	return nil
}

func (ctx *CmdCtx) LogConfigure(_ *cobra.Command, _ []string) (err error) {
	config := ctx.config.Log

	switch config.Level {
	case logging.LevelError, logging.LevelWarn, logging.LevelInfo, logging.LevelDebug, logging.LevelTrace:
		break
	default:
		config.Level = logging.LevelTrace
	}

	switch config.Format {
	case logging.FormatText, logging.FormatJSON:
		break
	default:
		config.Format = logging.FormatText
	}

	config.KeepStdout = true

	if err = logging.InitializeLogger(schema.Log{Level: ctx.config.Log.Level}, false); err != nil {
		return fmt.Errorf("Cannot initialize logger: %w", err)
	}

	ctx.log.WithFields(map[string]any{"filters": ctx.cconfig.filters, "files": ctx.cconfig.files}).Debug("Loaded Configuration Sources")
	ctx.log.WithFields(map[string]any{"level": ctx.config.Log.Level, "format": ctx.config.Log.Format, "file": ctx.config.Log.FilePath, "keep_stdout": ctx.config.Log.KeepStdout}).Debug("Logging Initialized")

	return nil
}

func (ctx *CmdCtx) LogProcessCurrentUserRunE(_ *cobra.Command, _ []string) (err error) {
	var current *user.User

	if current, err = user.Current(); err != nil {
		current = &user.User{Uid: strconv.Itoa(syscall.Getuid()), Gid: strconv.Itoa(syscall.Getgid())}
	}

	fields := map[string]any{"uid": current.Uid, "gid": current.Gid}

	if current.Username != "" {
		fields["username"] = current.Username
	}

	if current.Name != "" {
		fields["name"] = current.Name
	}

	var gids []string

	if gids, err = current.GroupIds(); err == nil && len(gids) != 0 {
		gidsFinal := []string{}

		for _, gid := range gids {
			if gid == current.Gid {
				continue
			}

			gidsFinal = append(gidsFinal, gid)
		}

		if len(gidsFinal) != 0 {
			fields["gids"] = strings.Join(gidsFinal, ",")
		}
	}

	ctx.log.WithFields(fields).Debug("Process user information")

	return nil
}

// ConfigValidateLogRunE logs the warnings and errors detected during the validations that have ran.
func (ctx *CmdCtx) ConfigValidateLogRunE(_ *cobra.Command, _ []string) (err error) {
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

// ConfigValidateSectionPasswordRunE validates the configuration (structure, password section).
func (ctx *CmdCtx) ConfigValidateSectionPasswordRunE(cmd *cobra.Command, _ []string) (err error) {
	if ctx.config.AuthenticationBackend.File == nil {
		return fmt.Errorf("password configuration was not initialized")
	}

	val := &schema.StructValidator{}

	validator.ValidatePasswordConfiguration(&ctx.config.AuthenticationBackend.File.Password, val)

	errs := val.Errors()

	if len(errs) == 0 {
		return nil
	}

	for i, e := range errs {
		if i == 0 {
			err = e
			continue
		}

		err = fmt.Errorf("%v, %w", err, e)
	}

	return fmt.Errorf("errors occurred validating the password configuration: %w", err)
}

// ConfigEnsureExistsRunE logs the warnings and errors detected during the validations that have ran.
func (ctx *CmdCtx) ConfigEnsureExistsRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		configs []string
		created bool
		result  XEnvCLIResult
	)

	if configs, result, err = loadXEnvCLIStringSliceValue(cmd, cmdFlagEnvNameConfig, cmdFlagNameConfig); err != nil {
		return err
	}

	switch {
	case result == XEnvCLIResultCLIExplicit:
		return nil
	case result == XEnvCLIResultEnvironment && len(configs) == 1:
		switch configs[0] {
		case cmdConfigDefaultContainer, cmdConfigDefaultDaemon:
			break
		default:
			return nil
		}
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

// HelperConfigLoadRunE loads the configuration into the CmdCtx.
func (ctx *CmdCtx) HelperConfigLoadRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		definitions *schema.Definitions
		filters     []configuration.BytesFilter
	)

	if ctx.cconfig == nil {
		ctx.cconfig = NewCmdCtxConfig()
	}

	if ctx.cconfig.files, filters, err = loadXEnvCLIConfigValues(cmd); err != nil {
		return err
	}

	ctx.cconfig.filters = make([]string, len(filters))

	for i, filter := range filters {
		ctx.cconfig.filters[i] = filter.Name()
	}

	ctx.cconfig.sources = configuration.NewDefaultSourcesWithDefaults(
		ctx.cconfig.files,
		filters,
		configuration.DefaultEnvPrefix,
		configuration.DefaultEnvDelimiter,
		ctx.cconfig.defaults,
		ctx.cconfig.sources...)

	if definitions, err = configuration.LoadDefinitions(ctx.cconfig.validator, ctx.cconfig.sources...); err != nil {
		return err
	}

	if ctx.cconfig.keys, err = configuration.LoadAdvanced(
		ctx.cconfig.validator,
		"",
		ctx.config,
		definitions,
		ctx.cconfig.sources...); err != nil {
		return err
	}

	return nil
}
