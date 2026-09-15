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

func isPasswordChangeRequired(ctx *middlewares.AutheliaCtx, username string) (required bool, err error) {
	name := ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute

	if name == "" {
		return false, nil
	}

	var details *authentication.UserDetailsExtended

	if details, err = ctx.Providers.UserProvider.GetDetailsExtended(username); err != nil {
		return false, fmt.Errorf("error occurred retrieving extended user details to determine if a password change is required: %w", err)
	}

	value, found := ctx.Providers.UserAttributeResolver.ResolveWithExtra(name, details, ctx.GetClock().Now(), details.Extra)
	if !found {
		return false, nil
	}

	if required, ok := value.(bool); ok {
		return required, nil
	}

	ctx.Logger.WithFields(map[string]any{"username": username, "attribute": name}).
		Errorf("The '%s' user attribute did not resolve to a boolean so it can't determine if a password change is required", name)

	return false, nil
}

func completePasswordChangeRequired(ctx *middlewares.AutheliaCtx, provider *session.Session, username string) (err error) {
	if err = clearPasswordChangeRequired(ctx, username); err != nil {
		return err
	}

	if err = provider.DestroySession(ctx.RequestCtx); err != nil {
		ctx.Logger.WithError(err).WithFields(map[string]any{"username": username}).
			Warn("Error occurred destroying the session which was held pending a required password change, resetting it instead")

		if err = provider.SaveSession(ctx.RequestCtx, provider.NewDefaultUserSession()); err != nil {
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

	if required, err = isPasswordChangeRequired(ctx, username); err != nil || !required {
		return err
	}

	if err = ctx.Providers.UserProvider.ClearExtraAttribute(username, name); err != nil {
		return fmt.Errorf("error occurred clearing the '%s' attribute which requires the user change their password: %w", name, err)
	}

	return nil
}
