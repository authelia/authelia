// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// ConfigurationGET get the configuration accessible to authenticated users.
func ConfigurationGET(ctx *middlewares.AutheliaCtx) {
	body := configurationBody{
		AvailableMethods:       make(MethodList, 0, 3),
		PasswordChangeDisabled: false,
		PasswordResetDisabled:  false,
	}

	// SecondFactorEnabled tells the portal whether signing in may require a second factor, which only the access
	// control rules and the OpenID Connect client authorization policies decide. AvailableMethods are the methods users can register and use, which also covers session
	// elevation and passkey logins.
	body.SecondFactorEnabled = ctx.Providers.Authorizer.IsSecondFactorEnabled()

	switch {
	case secondFactorUsable(ctx):
		body.AvailableMethods = ctx.AvailableSecondFactorMethods()
	case passkeyRegistrationUsable(ctx):
		// Nothing requires a second factor, but users still need a way to register the passkeys they sign in with.
		body.AvailableMethods = MethodList{model.SecondFactorMethodWebAuthn}
	}

	body.PasswordChangeDisabled = ctx.Configuration.AuthenticationBackend.PasswordChange.Disable
	body.PasswordResetDisabled = ctx.Configuration.AuthenticationBackend.PasswordReset.Disable
	body.RecoveryCodesDisabled = ctx.Configuration.RecoveryCodes.Disable

	ctx.Logger.WithFields(
		map[string]any{
			"available_methods":        body.AvailableMethods,
			"second_factor_enabled":    body.SecondFactorEnabled,
			"password_change_disabled": body.PasswordChangeDisabled,
			"password_reset_disabled":  body.PasswordResetDisabled,
			"recovery_codes_disabled":  body.RecoveryCodesDisabled,
		}).Trace("Authelia configuration requested")

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.Errorf("Unable to set configuration response in body: %s", err)
	}
}

func secondFactorUsable(ctx *middlewares.AutheliaCtx) bool {
	if ctx.Providers.Authorizer.IsSecondFactorEnabled() {
		return true
	}

	elevation := ctx.Configuration.IdentityValidation.ElevatedSession

	return elevation.RequireSecondFactor || elevation.SkipSecondFactor
}

// passkeyRegistrationUsable returns true when passkey logins are enabled and WebAuthn is usable from the request
// origin, in which case users must be able to register WebAuthn credentials even when no access control rule or
// elevation setting requires a second factor.
func passkeyRegistrationUsable(ctx *middlewares.AutheliaCtx) bool {
	if ctx.Configuration.WebAuthn.Disable || !ctx.Configuration.WebAuthn.EnablePasskeyLogin {
		return false
	}

	origin, err := ctx.GetOrigin()

	return err == nil && webauthn.IsOriginAvailable(ctx.Configuration.WebAuthn, origin)
}
