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

	body.SecondFactorEnabled = ctx.Providers.Authorizer.IsSecondFactorEnabled()

	switch {
	case secondFactorUsable(ctx):
		body.AvailableMethods = ctx.AvailableSecondFactorMethods()
	case passkeyRegistrationUsable(ctx):
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

func passkeyRegistrationUsable(ctx *middlewares.AutheliaCtx) bool {
	if ctx.Configuration.WebAuthn.Disable || !ctx.Configuration.WebAuthn.EnablePasskeyLogin {
		return false
	}

	origin, err := ctx.GetOrigin()

	return err == nil && webauthn.IsOriginAvailable(ctx.Configuration.WebAuthn, origin)
}
