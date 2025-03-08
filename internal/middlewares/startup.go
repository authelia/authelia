package middlewares

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

func (p *Providers) StartupChecks(ctx Context, log bool) (err error) {
	e := &ErrProviderStartupCheck{errors: map[string]error{}}

	var (
		disable  bool
		provider model.StartupCheck
	)

	provider, disable = ctx.GetProviders().StorageProvider, false
	doStartupCheck(ctx, ProviderNameStorage, provider, disable, log, e.errors)

	provider, disable = ctx.GetProviders().UserProvider, false
	doStartupCheck(ctx, ProviderNameUser, provider, disable, log, e.errors)

	provider, disable = ctx.GetProviders().Notifier, false
	doStartupCheck(ctx, ProviderNameNotification, provider, disable, log, e.errors)

	provider, disable = ctx.GetProviders().NTP, ctx.GetConfiguration().NTP.DisableStartupCheck
	doStartupCheck(ctx, ProviderNameNTP, provider, disable, log, e.errors)

	provider, disable = ctx.GetProviders().UserAttributeResolver, false
	doStartupCheck(ctx, ProviderNameExpressions, provider, disable, log, e.errors)

	provider = ctx.GetProviders().MetaDataService
	disable = !ctx.GetConfiguration().WebAuthn.Metadata.Enabled || ctx.GetProviders().MetaDataService == nil
	doStartupCheck(ctx, ProviderNameWebAuthnMetaData, provider, disable, log, e.errors)

	var filters []string

	if ctx.GetConfiguration().NTP.DisableFailure {
		filters = append(filters, ProviderNameNTP)
	}

	return e.FilterError(filters...)
}

func doStartupCheck(ctx Context, name string, provider model.StartupCheck, disabled, log bool, errors map[string]error) {
	if log {
		ctx.GetLogger().WithFields(map[string]any{LogFieldProvider: name}).Trace(LogMessageStartupCheckPerforming)
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

	var err error

	if err = provider.StartupCheck(); err != nil {
		if log {
			ctx.GetLogger().WithError(err).WithField(LogFieldProvider, name).Error(LogMessageStartupCheckError)
		}

		errors[name] = err

		return
	}

	if log {
		ctx.GetLogger().WithFields(map[string]any{LogFieldProvider: name}).Trace("Startup Check Completed Successfully")
	}
}

type ErrProviderStartupCheck struct {
	errors map[string]error
}

func (e *ErrProviderStartupCheck) Error() string {
	keys := make([]string, 0, len(e.errors))
	for k := range e.errors {
		keys = append(keys, k)
	}

	return fmt.Sprintf("errors occurred performing checks on the '%s' providers", strings.Join(keys, ", "))
}

func (e *ErrProviderStartupCheck) Failed() (failed []string) {
	for key := range e.errors {
		failed = append(failed, key)
	}

	return failed
}

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

func (e *ErrProviderStartupCheck) ErrorMap() map[string]error {
	return e.errors
}
