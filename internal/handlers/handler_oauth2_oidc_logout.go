// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func oidcLogout(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, pending *model.OAuth2SessionLogout) {
	if ctx.Configuration.IdentityProviders.OIDC == nil {
		return
	}

	if !oidcLogoutAuthenticated(userSession) {
		return
	}

	provider, err := ctx.GetSessionProvider()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred obtaining the session provider during logout")

		return
	}

	issuer := provider.GetIssuer()

	sid := oidcLogoutEndedSessionID(ctx, issuer, userSession, pending)

	oidcLogoutSession(ctx, issuer, userSession.PublicID)

	if sid != "" {
		oidcLogoutEndedSession(ctx, userSession.Username, pending)
	}
}

func oidcLogoutAuthenticated(userSession *session.UserSession) bool {
	return userSession != nil && userSession.Username != "" && userSession.PublicID != ""
}

func oidcLogoutEndedSessionID(ctx *middlewares.AutheliaCtx, issuer string, userSession *session.UserSession, pending *model.OAuth2SessionLogout) (sid string) {
	if pending == nil || pending.SessionID == "" {
		return ""
	}

	record, err := ctx.Providers.StorageProvider.LoadOAuth2SessionIDBySessionID(ctx, issuer, pending.SessionID)

	switch {
	case err != nil:
		ctx.GetLogger().WithError(err).Error("Error occurred loading the session identified by the 'id_token_hint' during logout")

		return ""
	case record == nil:
		return pending.SessionID
	case record.PublicID != userSession.PublicID:
		ctx.GetLogger().Debugf("The session identified by the 'id_token_hint' is not the session of user '%s' which is being logged out instead", userSession.Username)
	}

	return ""
}

func oidcLogoutUsername(ctx *middlewares.AutheliaCtx, subject string) (username string) {
	id, err := uuid.Parse(subject)
	if err != nil {
		return ""
	}

	var identifier *model.UserOpaqueIdentifier

	if identifier, err = ctx.Providers.StorageProvider.LoadUserOpaqueIdentifier(ctx, id); err != nil || identifier == nil {
		ctx.GetLogger().WithError(err).Debug("The user of the subject of the 'id_token_hint' could not be determined during logout")

		return ""
	}

	return identifier.Username
}

func oidcLogoutSession(ctx *middlewares.AutheliaCtx, issuer, publicID string) {
	var (
		sids []string
		err  error
	)

	for range oidcLogoutEndAttempts {
		if sids, err = ctx.Providers.StorageProvider.EndOAuth2Session(ctx, issuer, publicID, oidc.NewLogoutRevokeSchedule(ctx.Configuration.IdentityProviders.OIDC, ctx.GetClock().Now())); err == nil {
			break
		}
	}

	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred ending the OpenID Connect sessions during logout: they'll be ended once it's determined the session no longer exists")

		return
	}

	for _, sid := range sids {
		oidcLogoutRevoke(ctx, sid, true)
	}
}

func oidcLogoutEndedSession(ctx *middlewares.AutheliaCtx, username string, pending *model.OAuth2SessionLogout) {
	if !strings.EqualFold(oidcLogoutUsername(ctx, pending.Subject), username) {
		ctx.GetLogger().Debugf("The ended session identified by the 'id_token_hint' was skipped during logout: it doesn't belong to user '%s'", username)

		return
	}

	schedule := oidc.NewLogoutRevokeSchedule(ctx.Configuration.IdentityProviders.OIDC, ctx.GetClock().Now())

	var err error

	for range oidcLogoutEndAttempts {
		if err = ctx.Providers.StorageProvider.SaveOAuth2SessionLogoutTask(ctx, schedule.RevokeTask(pending.SessionID)); err == nil {
			break
		}
	}

	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred ending the OpenID Connect session identified by the 'id_token_hint' during logout")
	}

	oidcLogoutRevoke(ctx, pending.SessionID, err == nil)
}

func oidcLogoutRevoke(ctx *middlewares.AutheliaCtx, sid string, recorded bool) {
	err := ctx.Providers.StorageProvider.RevokeOAuth2RefreshTokenSessionsBySessionID(ctx, sid)

	switch {
	case err == nil:
		return
	case recorded:
		ctx.GetLogger().WithError(err).Error("Error occurred revoking the refresh tokens of an ended OpenID Connect session during logout: they'll be revoked shortly")
	default:
		ctx.GetLogger().WithError(err).Errorf("Error occurred revoking the refresh tokens of an ended OpenID Connect session during logout: they won't be revoked as the end of the session with id '%s' could not be recorded, and its online refresh tokens remain usable until they expire", sid)
	}
}

const oidcLogoutEndAttempts = 3

var (
	errLogoutHintAnotherSession = errors.New("the 'id_token_hint' identifies a session other than the one which is logged in")
	errLogoutHintAnotherUser    = errors.New("the 'id_token_hint' was not issued to the currently authenticated End-User")
)

func oidcLogoutValidate(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, pending *model.OAuth2SessionLogout) (err error) {
	if !oidcLogoutAuthenticated(userSession) {
		return nil
	}

	if pending.SessionID != "" {
		var provider session.Strategy

		if provider, err = ctx.GetSessionProvider(); err != nil {
			return fmt.Errorf("error obtaining the session provider: %w", err)
		}

		var record *model.OAuth2SessionID

		if record, err = ctx.Providers.StorageProvider.LoadOAuth2SessionIDBySessionID(ctx, provider.GetIssuer(), pending.SessionID); err != nil {
			return fmt.Errorf("error loading the session identified by the 'id_token_hint': %w", err)
		}

		if record != nil {
			if record.PublicID != userSession.PublicID {
				return errLogoutHintAnotherSession
			}

			return nil
		}
	}

	if pending.Subject == "" {
		return nil
	}

	if !strings.EqualFold(oidcLogoutUsername(ctx, pending.Subject), userSession.Username) {
		return errLogoutHintAnotherUser
	}

	return nil
}
