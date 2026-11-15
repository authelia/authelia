// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package embed

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/service"
)

// ProvidersRetainer is implemented by a Context which can retain the providers completed by ProvidersStartupCheck. The
// session providers don't exist until then, so a Context returned by NewContext implements it, and a custom Context must
// as well.
type ProvidersRetainer interface {
	SetProviders(providers middlewares.Providers)
}

// ProvidersStartupCheck performs the startup checks for all providers given a context, then completes the providers
// which depend on the others and retains them in the context. It must be called before ServiceRunAll.
func ProvidersStartupCheck(ctx Context, log bool) (err error) {
	providers := ctx.GetProviders()

	if err = providers.StartupChecks(ctx, log); err != nil {
		return err
	}

	retainer, ok := ctx.(ProvidersRetainer)
	if !ok {
		return fmt.Errorf("the context does not support retaining the finalized providers as it does not implement SetProviders")
	}

	if err = providers.Finalize(ctx); err != nil {
		return err
	}

	retainer.SetProviders(providers)

	return nil
}

// ServiceRunAll runs all services given a context.
func ServiceRunAll(ctx Context) (err error) {
	if ctx == nil {
		return fmt.Errorf("no context provided")
	}

	if ctx.GetConfiguration() == nil {
		return fmt.Errorf("no configuration provided")
	}

	if ctx.GetLogger() == nil {
		return fmt.Errorf("no logger provided")
	}

	return service.RunAll(ctx)
}
