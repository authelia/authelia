// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

func isPasswordChangeRequired(ctx AuthzContext, username string, cached bool) (required bool, err error) {
	if ctx.GetConfiguration().AuthenticationBackend.PasswordChange.RequiredAttribute == "" {
		return false, nil
	}

	get := ctx.GetUserProvider().GetDetailsExtended

	if cached {
		get = ctx.GetUserProvider().GetDetailsExtendedCached
	}

	var details *authentication.UserDetailsExtended

	if details, err = get(username); err != nil {
		return false, fmt.Errorf("error occurred retrieving extended user details to determine if a password change is required: %w", err)
	}

	return isPasswordChangeRequiredDetails(ctx, details), nil
}

func isPasswordChangeRequiredDetails(ctx AuthzContext, details *authentication.UserDetailsExtended) (required bool) {
	name := ctx.GetConfiguration().AuthenticationBackend.PasswordChange.RequiredAttribute

	if name == "" || details == nil {
		return false
	}

	value, found := ctx.GetProviderUserAttributeResolver().ResolveWithExtra(name, details, ctx.GetClock().Now(), details.Extra)
	if !found {
		return false
	}

	if required, ok := value.(bool); ok {
		return required
	}

	ctx.GetLogger().WithFields(map[string]any{"username": details.GetUsername(), "attribute": name}).
		Errorf("The '%s' user attribute did not resolve to a boolean so it can't determine if a password change is required", name)

	return false
}

func getFirstFactorUserDetails(ctx *middlewares.AutheliaCtx, username string) (details *authentication.UserDetails, extended *authentication.UserDetailsExtended, err error) {
	if ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute == "" {
		details, err = ctx.Providers.UserProvider.GetDetails(username)

		return details, nil, err
	}

	if extended, err = ctx.Providers.UserProvider.GetDetailsExtended(username); err != nil || extended == nil {
		return nil, nil, err
	}

	return extended.UserDetails, extended, nil
}

func setPasswordChangeRequiredSession(ctx *middlewares.AutheliaCtx, provider session.Strategy, username string) (userSession session.UserSession) {
	userSession = provider.NewDefault()

	userSession.SetPasswordChangeRequired(ctx.GetClock().Now(), username)

	ctx.GetLogger().WithFields(map[string]any{"username": username}).
		Info("User authenticated but the backend requires them to change their password, so the session is held pending that change")

	return userSession
}

func completePasswordChangeRequired(ctx *middlewares.AutheliaCtx, provider session.Strategy, username string) (err error) {
	if err = clearPasswordChangeRequired(ctx, username); err != nil {
		ctx.GetLogger().WithError(err).WithFields(map[string]any{"username": username, "attribute": ctx.Configuration.AuthenticationBackend.PasswordChange.ClearAttribute}).
			Error("Error occurred clearing the attribute which requires the user change their password, the password was changed so the user may sign in with it")
	}

	if err = provider.Destroy(ctx); err != nil {
		ctx.GetLogger().WithError(err).WithFields(map[string]any{"username": username}).
			Warn("Error occurred destroying the session which was held pending a required password change, resetting it instead")

		userSession := provider.NewDefault()

		if err = provider.Save(ctx, &userSession); err != nil {
			return fmt.Errorf("error occurred resetting the session which was held pending a required password change: %w", err)
		}
	}

	return nil
}

func clearPasswordChangeRequired(ctx *middlewares.AutheliaCtx, username string) (err error) {
	name := ctx.Configuration.AuthenticationBackend.PasswordChange.ClearAttribute

	if name == "" {
		return nil
	}

	var required bool

	if required, err = isPasswordChangeRequired(ctx, username, false); err != nil || !required {
		return err
	}

	if err = ctx.Providers.UserProvider.ClearExtraAttribute(username, name); err != nil {
		return fmt.Errorf("error occurred clearing the '%s' attribute which requires the user change their password: %w", name, err)
	}

	return nil
}
