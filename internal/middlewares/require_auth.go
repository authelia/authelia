// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

// Require1FA check if user has enough permissions to execute the next handler.
func Require1FA(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if s, err := ctx.GetSession(); err != nil || s.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA) < authentication.OneFactor {
			ctx.ReplyForbidden()
			return
		}

		next(ctx)
	}
}

// RequireElevated requires various elevation criteria.
func RequireElevated(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		var (
			userSession session.UserSession
			err         error
		)
		if userSession, err = ctx.GetSession(); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred attempting to lookup user session during an elevation check.")

			if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{FirstFactor: true}}, fasthttp.StatusForbidden); err != nil {
				ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during an elevation check.")
			}

			return
		}

		if userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA) < authentication.OneFactor {
			ctx.Logger.Warn("An anonymous user attempted to access an elevated protected endpoint.")

			if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{FirstFactor: true}}, fasthttp.StatusForbidden); err != nil {
				ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during an elevation check.")
			}

			return
		}

		if !handleRequireElevatedShouldDoNext(ctx, &userSession) {
			return
		}

		next(ctx)
	}
}

// RequireElevationSecondFactor requires one factor authentication, and second factor authentication as well where the
// elevated session policy would require it of the user, without requiring an elevated session.
func RequireElevationSecondFactor(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		var (
			userSession session.UserSession
			level       authentication.Level
			err         error
		)

		if userSession, err = ctx.GetSession(); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred attempting to lookup user session during a second factor check.")

			if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{FirstFactor: true}}, fasthttp.StatusForbidden); err != nil {
				ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during a second factor check.")
			}

			return
		}

		level = userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

		if level < authentication.OneFactor {
			ctx.Logger.Warn("An anonymous user attempted to access a second factor protected endpoint.")

			if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{FirstFactor: true}}, fasthttp.StatusForbidden); err != nil {
				ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during a second factor check.")
			}

			return
		}

		if !handleRequireSecondFactorShouldDoNext(ctx, &userSession, level) {
			return
		}

		next(ctx)
	}
}

func handleRequireElevatedShouldDoNext(ctx *AutheliaCtx, userSession *session.UserSession) (doNext bool) {
	var err error

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if ctx.Configuration.IdentityValidation.ElevatedSession.SkipSecondFactor && level >= authentication.TwoFactor {
		ctx.Logger.WithFields(map[string]any{"user": userSession.Username}).Trace("The user session elevation was not checked as the user has performed second factor authentication and the policy to skip this is enabled.")

		return true
	}

	if !handleRequireSecondFactorShouldDoNext(ctx, userSession, level) {
		return
	}

	if userSession.Elevations.User == nil {
		if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{Elevation: true}}, fasthttp.StatusForbidden); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during an elevation check.")
		}

		return
	}

	return handleRequireElevatedShouldDoNextValidate(ctx, userSession)
}

func handleRequireSecondFactorShouldDoNext(ctx *AutheliaCtx, userSession *session.UserSession, level authentication.Level) (doNext bool) {
	var (
		info model.UserInfo
		err  error
	)

	if !ctx.Configuration.IdentityValidation.ElevatedSession.RequireSecondFactor || level >= authentication.TwoFactor {
		return true
	}

	if info, err = ctx.Providers.StorageProvider.LoadUserInfo(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred attempting to lookup user information during a elevation check.")

		if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{SecondFactor: true}}, fasthttp.StatusForbidden); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during an elevation check.")
		}

		return
	}

	if info.HasTOTP || info.HasWebAuthn || info.HasDuo {
		ctx.Logger.WithFields(map[string]any{"user": userSession.Username}).Info("The user session elevation was not checked as the user must have also performed second factor authentication.")

		if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{SecondFactor: true}}, fasthttp.StatusForbidden); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during an elevation check.")
		}

		return
	}

	return true
}

func handleRequireElevatedShouldDoNextValidate(ctx *AutheliaCtx, userSession *session.UserSession) (doNext bool) {
	var err error

	invalid := false

	if ctx.GetClock().Now().After(userSession.Elevations.User.Expires) {
		invalid = true

		ctx.Logger.WithFields(map[string]any{"user": userSession.Username, "expired": userSession.Elevations.User.Expires.Unix()}).Info("The user session elevation was expired. It will be destroyed and the users access will be forbidden.")
	}

	if !ctx.RemoteIP().Equal(userSession.Elevations.User.RemoteIP) {
		invalid = true

		ctx.Logger.WithFields(map[string]any{"user": userSession.Username, "expected_ip": userSession.Elevations.User.RemoteIP.String()}).Warn("The user session elevation did not have a matching IP. It will be destroyed and the users access will be forbidden.")
	}

	if invalid {
		userSession.Elevations.User = nil

		if err = ctx.SaveSession(*userSession); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred trying to save the user session after a policy constraint violation occurred.")
		}

		if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{Elevation: true}}, fasthttp.StatusForbidden); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during an elevation check.")
		}

		return
	}

	return true
}
