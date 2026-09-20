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

// RequireElevatedOrPasswordChangeRequired is RequireElevated which additionally admits a session held at sign in
// pending a required password change.
//
// Such a session deliberately holds no authentication level, so RequireElevated alone would refuse it and the user
// would have no way to perform the change they are being held for. It is applied to the password change endpoint
// only: it must never guard the credential registration endpoints, where a held session would let somebody with an
// administrator issued password enroll or remove a second factor without ever completing authentication.
func RequireElevatedOrPasswordChangeRequired(next RequestHandler) RequestHandler {
	elevated := RequireElevated(next)

	return func(ctx *AutheliaCtx) {
		userSession, err := ctx.GetSession()
		if err == nil && userSession.IsPasswordChangeRequired() {
			next(ctx)

			return
		}

		elevated(ctx)
	}
}

func handleRequireElevatedShouldDoNext(ctx *AutheliaCtx, userSession *session.UserSession) (doNext bool) {
	var err error

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if ctx.Configuration.IdentityValidation.ElevatedSession.SkipSecondFactor && level >= authentication.TwoFactor {
		ctx.Logger.WithFields(map[string]any{"user": userSession.Username}).Trace("The user session elevation was not checked as the user has performed second factor authentication and the policy to skip this is enabled.")

		return true
	}

	if ctx.Configuration.IdentityValidation.ElevatedSession.RequireSecondFactor && level < authentication.TwoFactor {
		var info model.UserInfo

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
	}

	if userSession.Elevations.User == nil {
		if err = ctx.ReplyJSON(OKResponse{Status: "KO", Data: ElevatedForbiddenResponse{Elevation: true}}, fasthttp.StatusForbidden); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred encoding JSON response during an elevation check.")
		}

		return
	}

	return handleRequireElevatedShouldDoNextValidate(ctx, userSession)
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
