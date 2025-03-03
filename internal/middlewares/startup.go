package middlewares

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/model"
)

func (p *Providers) StartupChecks(ctx Context) {
	var (
		failures []string
		err      error
	)

	ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameStorage}).Trace(logMessageStartupCheckPerforming)

	if err = doStartupCheck(ctx, providerNameStorage, ctx.GetProviders().StorageProvider, false); err != nil {
		ctx.GetLogger().WithError(err).WithField(logFieldProvider, providerNameStorage).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameStorage)
	} else {
		ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameStorage}).Trace(logMessageStartupCheckSuccess)
	}

	ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameUser}).Trace(logMessageStartupCheckPerforming)

	if err = doStartupCheck(ctx, providerNameUser, ctx.GetProviders().UserProvider, false); err != nil {
		ctx.GetLogger().WithError(err).WithField(logFieldProvider, providerNameUser).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameUser)
	} else {
		ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameUser}).Trace(logMessageStartupCheckSuccess)
	}

	ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameNotification}).Trace(logMessageStartupCheckPerforming)

	if err = doStartupCheck(ctx, providerNameNotification, ctx.GetProviders().Notifier, ctx.GetConfiguration().Notifier.DisableStartupCheck); err != nil {
		ctx.GetLogger().WithError(err).WithField(logFieldProvider, providerNameNotification).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameNotification)
	} else {
		ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameNotification}).Trace(logMessageStartupCheckSuccess)
	}

	ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameNTP}).Trace(logMessageStartupCheckPerforming)

	if err = doStartupCheck(ctx, providerNameNTP, ctx.GetProviders().NTP, ctx.GetConfiguration().NTP.DisableStartupCheck); err != nil {
		if !ctx.GetConfiguration().NTP.DisableFailure {
			ctx.GetLogger().WithError(err).WithField(logFieldProvider, providerNameNTP).Error(logMessageStartupCheckError)

			failures = append(failures, providerNameNTP)
		} else {
			ctx.GetLogger().WithError(err).WithField(logFieldProvider, providerNameNTP).Warn(logMessageStartupCheckError)
		}
	} else {
		ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameNTP}).Trace(logMessageStartupCheckSuccess)
	}

	ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameExpressions}).Trace(logMessageStartupCheckPerforming)

	if err = doStartupCheck(ctx, providerNameExpressions, ctx.GetProviders().UserAttributeResolver, false); err != nil {
		ctx.GetLogger().WithError(err).WithField(logFieldProvider, providerNameExpressions).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameExpressions)
	} else {
		ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameExpressions}).Trace(logMessageStartupCheckSuccess)
	}

	if err = doStartupCheck(ctx, providerNameWebAuthnMetaData, ctx.GetProviders().MetaDataService, !ctx.GetConfiguration().WebAuthn.Metadata.Enabled || ctx.GetProviders().MetaDataService == nil); err != nil {
		ctx.GetLogger().WithError(err).WithField(logFieldProvider, providerNameWebAuthnMetaData).Error(logMessageStartupCheckError)

		failures = append(failures, providerNameWebAuthnMetaData)
	} else {
		ctx.GetLogger().WithFields(map[string]any{logFieldProvider: providerNameWebAuthnMetaData}).Trace("Startup Check Completed Successfully")
	}

	if len(failures) != 0 {
		ctx.GetLogger().WithField("providers", failures).Fatalf("One or more providers had fatal failures performing startup checks, for more detail check the error level logs")
	}
}

func doStartupCheck(ctx Context, name string, provider model.StartupCheck, disabled bool) error {
	if disabled {
		ctx.GetLogger().Debugf("%s provider: startup check skipped as it is disabled", name)
		return nil
	}

	if provider == nil {
		return fmt.Errorf("unrecognized provider or it is not configured properly")
	}

	return provider.StartupCheck()
}
