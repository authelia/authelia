package embed

func ProvidersStartupCheck(ctx Context, log bool) (err error) {
	providers := ctx.GetProviders()

	return providers.StartupChecks(ctx, log)
}
