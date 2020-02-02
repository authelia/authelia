package handlers

import (
	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/middlewares"
)

type ExtendedConfigurationBody struct {
	AvailableMethods MethodList `json:"available_methods"`

	// OneFactorDefaultPolicy is set if default policy is 'one_factor'
	OneFactorDefaultPolicy bool `json:"one_factor_default_policy"`
}

// ExtendedConfigurationGet get the extended configuration accessible to authenticated users.
func ExtendedConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ExtendedConfigurationBody{}
	body.AvailableMethods = MethodList{authentication.TOTP, authentication.U2F}

	if ctx.Configuration.DuoAPI != nil {
		body.AvailableMethods = append(body.AvailableMethods, authentication.Push)
	}

	defaultPolicy := authorization.PolicyToLevel(ctx.Configuration.AccessControl.DefaultPolicy)
	body.OneFactorDefaultPolicy = defaultPolicy == authorization.OneFactor
	ctx.Logger.Tracef("Default policy set to one factor: %v", body.OneFactorDefaultPolicy)

	ctx.Logger.Tracef("Available methods are %s", body.AvailableMethods)
	ctx.SetJSONBody(body)
}
