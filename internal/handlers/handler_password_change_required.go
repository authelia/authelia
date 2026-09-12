// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func isPasswordChangeRequired(ctx *middlewares.AutheliaCtx, username string) (required bool) {
	name := ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute

	if name == "" {
		return false
	}

	var (
		details *authentication.UserDetailsExtended
		err     error
	)

	if details, err = ctx.Providers.UserProvider.GetDetailsExtended(username); err != nil {
		ctx.Logger.WithError(err).WithFields(map[string]any{"username": username, "attribute": name}).
			Error("Error occurred retrieving extended user details to determine if a password change is required")

		return false
	}

	value, found := ctx.Providers.UserAttributeResolver.ResolveWithExtra(name, details, ctx.GetClock().Now(), details.Extra)
	if !found {
		return false
	}

	if required, ok := value.(bool); ok {
		return required
	}

	ctx.Logger.WithFields(map[string]any{"username": username, "attribute": name}).
		Errorf("The '%s' user attribute did not resolve to a boolean so it can't determine if a password change is required", name)

	return false
}

func clearPasswordChangeRequired(ctx *middlewares.AutheliaCtx, username string) {
	name := ctx.Configuration.AuthenticationBackend.PasswordChange.ClearAttribute

	if name == "" || !isPasswordChangeRequired(ctx, username) {
		return
	}

	if err := ctx.Providers.UserProvider.ClearExtraAttribute(username, name); err != nil {
		ctx.Logger.WithError(err).WithFields(map[string]any{"username": username, "attribute": name}).
			Error("Error occurred clearing the attribute which requires the user change their password")
	}
}
