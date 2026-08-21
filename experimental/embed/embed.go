package embed

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/service"
)

// ProvidersStartupCheck performs the startup checks for all providers given a context.
func ProvidersStartupCheck(ctx Context, log bool) (err error) {
	providers := ctx.GetProviders()

	return providers.StartupChecks(ctx, log)
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
