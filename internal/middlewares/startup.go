package middlewares

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

// StartupChecks performs the startup checks for each of the providers.
func (p *Providers) StartupChecks(ctx ServiceContext, log bool) (err error) {
	config := ctx.GetConfiguration()

	if !config.Server.DisableHealthcheck {
		if err = writeHealthCheckEnvConfig(config); err != nil {
			return err
		}
	}

	e := &ErrProviderStartupCheck{errors: map[string]error{}}

	var (
		disable  bool
		provider model.StartupCheck
	)

	provider, disable = ctx.GetProviders().StorageProvider, false
	doStartupCheck(ctx, ProviderNameStorage, provider, nil, disable, log, e.errors)

	provider, disable = ctx.GetProviders().Cache, false
	doStartupCheck(ctx, ProviderNameCache, provider, nil, disable, log, e.errors)

	provider, disable = ctx.GetProviders().UserProvider, false
	doStartupCheck(ctx, ProviderNameUser, provider, nil, disable, log, e.errors)

	provider, disable = ctx.GetProviders().Notifier, ctx.GetConfiguration().Notifier.DisableStartupCheck
	doStartupCheck(ctx, ProviderNameNotification, provider, nil, disable, log, e.errors)

	provider, disable = ctx.GetProviders().NTP, ctx.GetConfiguration().NTP.DisableStartupCheck
	doStartupCheck(ctx, ProviderNameNTP, provider, nil, disable, log, e.errors)

	provider, disable = ctx.GetProviders().UserAttributeResolver, false
	doStartupCheck(ctx, ProviderNameExpressions, provider, nil, disable, log, e.errors)

	provider = ctx.GetProviders().MetaDataService
	disable = !ctx.GetConfiguration().WebAuthn.Metadata.Enabled || ctx.GetProviders().MetaDataService == nil
	doStartupCheck(ctx, ProviderNameWebAuthnMetaData, provider, []string{ProviderNameStorage}, disable, log, e.errors)

	var filters []string

	if ctx.GetConfiguration().NTP.DisableFailure {
		filters = append(filters, ProviderNameNTP)
	}

	return e.FilterError(filters...)
}

func (p *Providers) Finalize(ctx ServiceContext) (err error) {
	var (
		repository  session.Repository
		sessionHMAC []byte
	)

	switch name := ctx.GetConfiguration().Session.Storage; name {
	case "internal":
		repository = storage.NewSessionRepository(p.StorageProvider)
	case "cache", "":
		repository = cache.NewSessionRepository(p.Cache)
	default:
		return fmt.Errorf("unknown session storage '%s'", name)
	}

	if sessionHMAC, err = p.StorageProvider.LoadHMACKey(ctx, "session", sha256.BlockSize); err != nil {
		return err
	}

	if p.Session, err = session.NewProvider(ctx.GetConfiguration(), sessionHMAC, p.Clock, p.Random, repository); err != nil {
		return err
	}

	p.SessionRepository = repository

	return nil
}

func doStartupCheck(ctx ServiceContext, name string, provider model.StartupCheck, required []string, disabled, log bool, errors map[string]error) {
	if log {
		ctx.GetLogger().WithFields(map[string]any{logging.FieldProvider: name}).Trace(LogMessageStartupCheckPerforming)
	}

	if disabled {
		if log {
			ctx.GetLogger().Debugf("%s provider: startup check skipped as it is disabled", name)
		}

		return
	}

	if provider == nil {
		errors[name] = fmt.Errorf("unrecognized provider or it is not configured properly")

		return
	}

	if len(required) > 0 {
		for _, rname := range required {
			if _, ok := errors[rname]; ok {
				err := fmt.Errorf("provider requires that the '%s' provider was successful but it wasn't", rname)

				errors[name] = err

				if log {
					ctx.GetLogger().WithError(err).WithField(logging.FieldProvider, name).Error(LogMessageStartupCheckError)
				}

				return
			}
		}
	}

	var err error
	if err = provider.StartupCheck(); err != nil {
		if log {
			ctx.GetLogger().WithError(err).WithField(logging.FieldProvider, name).Error(LogMessageStartupCheckError)
		}

		errors[name] = err

		return
	}

	if log {
		ctx.GetLogger().WithFields(map[string]any{logging.FieldProvider: name}).Trace("Startup Check Completed Successfully")
	}
}

func writeHealthCheckEnvConfig(config *schema.Configuration) (err error) {
	scheme := strProtoHTTP

	if config.Server.TLS.Certificate != "" && config.Server.TLS.Key != "" {
		scheme = strProtoHTTPS
	}

	host := config.Server.Address.Hostname()

	path := config.Server.Address.RouterPath()

	port := config.Server.Address.Port()

	return writeHealthCheckEnv(scheme, host, path, port)
}

func writeHealthCheckEnv(scheme, host, path string, port uint16) (err error) {
	if _, err = os.Stat("/app/healthcheck.sh"); err != nil {
		return nil
	}

	if _, err = os.Stat("/app/.healthcheck.env"); err != nil {
		return nil
	}

	var file *os.File

	if file, err = os.OpenFile("/app/.healthcheck.env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755); err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	if host == "0.0.0.0" {
		host = localhost
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}

	if path == "/" {
		path = ""
	}

	_, err = fmt.Fprintf(file, healthCheckEnv, scheme, host, port, path)

	return err
}

// ErrProviderStartupCheck is an error which contains the startup check error for each provider which failed.
type ErrProviderStartupCheck struct {
	errors map[string]error
}

// Error returns the string representation of this error.
func (e *ErrProviderStartupCheck) Error() string {
	keys := make([]string, 0, len(e.errors))
	for k := range e.errors {
		keys = append(keys, k)
	}

	return fmt.Sprintf("errors occurred performing checks on the '%s' providers", strings.Join(keys, ", "))
}

// Failed returns the names of the providers which failed their startup check.
func (e *ErrProviderStartupCheck) Failed() (failed []string) {
	for key := range e.errors {
		failed = append(failed, key)
	}

	return failed
}

// FilterError returns an error containing only the failures for the given providers, or nil if none of them failed.
func (e *ErrProviderStartupCheck) FilterError(providers ...string) error {
	filtered := map[string]error{}

	for provider, err := range e.errors {
		if utils.IsStringInSlice(provider, providers) {
			continue
		}

		filtered[provider] = err
	}

	if len(filtered) == 0 {
		return nil
	}

	return &ErrProviderStartupCheck{errors: filtered}
}

// ErrorMap returns the error for each provider which failed its startup check.
func (e *ErrProviderStartupCheck) ErrorMap() map[string]error {
	return e.errors
}
