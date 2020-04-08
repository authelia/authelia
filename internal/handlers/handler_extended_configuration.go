package handlers

import (
	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/middlewares"
)

// ExtendedConfigurationBody the content returned by extended configuration endpoint
type ExtendedConfigurationBody struct {
	AvailableMethods    MethodList `json:"available_methods"`
	SecondFactorEnabled bool       `json:"second_factor_enabled"` // whether second factor is enabled or not
	TOTPPeriod          int        `json:"totp_period"`
}

// ExtendedConfigurationGet get the extended configuration accessible to authenticated users.
func ExtendedConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ExtendedConfigurationBody{}
	body.AvailableMethods = MethodList{authentication.TOTP, authentication.U2F}
	body.TOTPPeriod = ctx.Configuration.TOTP.Period

	if ctx.Configuration.DuoAPI != nil {
		body.AvailableMethods = append(body.AvailableMethods, authentication.Push)
	}

	body.SecondFactorEnabled = ctx.Providers.Authorizer.IsSecondFactorEnabled()
	ctx.Logger.Tracef("Second factor enabled: %v", body.SecondFactorEnabled)

	ctx.Logger.Tracef("Available methods are %s", body.AvailableMethods)
	ctx.SetJSONBody(body)
}
