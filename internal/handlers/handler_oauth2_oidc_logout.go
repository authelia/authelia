// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

// oidcLogout ends the OpenID Connect sessions of the authenticated End-User. A logout requested by a user agent which
// is not authenticated only honors the post logout redirect, as the 'id_token_hint' alone doesn't establish that the
// user agent belongs to the End-User it was issued to.
func oidcLogout(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, pending *model.OAuth2LogoutSession) {
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

	oidcLogoutSession(ctx, issuer, userSession.Username, userSession.PublicID)

	if sid != "" {
		oidcLogoutEndedSession(ctx, userSession.Username, pending)
	}
}

// oidcLogoutAuthenticated returns true when the session belongs to an End-User, which is the only session a logout
// may end.
func oidcLogoutAuthenticated(userSession *session.UserSession) bool {
	return userSession != nil && userSession.Username != "" && userSession.PublicID != ""
}

// oidcLogoutEndedSessionID returns the session identifier of the 'id_token_hint' when the OpenID Connect session it
// identifies has already ended, which leaves the tokens issued to it to be revoked.
func oidcLogoutEndedSessionID(ctx *middlewares.AutheliaCtx, issuer string, userSession *session.UserSession, pending *model.OAuth2LogoutSession) (sid string) {
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

// oidcLogoutUsername returns the username the pairwise subject belongs to, or an empty string when it can't be
// determined.
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

// oidcLogoutSession ends every OpenID Connect session belonging to the Authelia session.
func oidcLogoutSession(ctx *middlewares.AutheliaCtx, issuer, username, publicID string) {
	// Delivered before the session identifiers are removed as the participating clients are recorded against them.
	oidcBackChannelLogoutSession(ctx, issuer, username, publicID)

	records, err := ctx.Providers.StorageProvider.LoadOAuth2SessionIDsByPublicID(ctx, issuer, publicID)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred loading the OpenID Connect session identifiers during logout")
	}

	for _, record := range records {
		if err = ctx.Providers.StorageProvider.RevokeOAuth2SessionsBySessionID(ctx, record.SessionID.String()); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred revoking the OAuth 2.0 sessions of an OpenID Connect session during logout")
		}
	}

	if err = ctx.Providers.StorageProvider.DeleteOAuth2SessionIDByPublicID(ctx, issuer, publicID); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred removing the OpenID Connect session identifiers during logout")
	}
}

// oidcLogoutEndedSession ends the OpenID Connect session identified by the 'id_token_hint' which outlived the Authelia
// session it belonged to. The participating clients were recorded against the Authelia session, so only the client
// which requested the logout is notified, though the tokens of every client issued to the session are revoked.
func oidcLogoutEndedSession(ctx *middlewares.AutheliaCtx, username string, pending *model.OAuth2LogoutSession) {
	if !strings.EqualFold(oidcLogoutUsername(ctx, pending.Subject), username) {
		ctx.GetLogger().Debugf("The ended session identified by the 'id_token_hint' was skipped during logout: it doesn't belong to user '%s'", username)

		return
	}

	if ctx.Providers.OpenIDConnect != nil && pending.ClientID != "" {
		if client, err := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, pending.ClientID); err != nil {
			ctx.GetLogger().WithError(err).Debugf("Back-Channel Logout request for user '%s' on client with id '%s' was skipped: the client is not registered", username, pending.ClientID)
		} else if client.GetBackChannelLogoutURI() != "" {
			oidcBackChannelLogoutSendGroups(ctx, username, []oidcBackChannelLogoutGroup{{sid: pending.SessionID, sectorID: client.GetSectorIdentifierURI(), subject: pending.Subject, clients: []oauthelia2.Client{client}}})
		}
	}

	if err := ctx.Providers.StorageProvider.RevokeOAuth2SessionsBySessionID(ctx, pending.SessionID); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred revoking the OAuth 2.0 sessions of the ended OpenID Connect session during logout")
	}
}

var (
	errLogoutHintAnotherSession = errors.New("the 'id_token_hint' identifies a session other than the one which is logged in")
	errLogoutHintAnotherUser    = errors.New("the 'id_token_hint' was not issued to the currently authenticated End-User")
)

// oidcLogoutValidate declines a logout request whose 'id_token_hint' identifies a session or End-User other than the one
// authenticated, as OpenID Connect RP-Initiated Logout 1.0 Section 2 permits of a hint which doesn't correspond to the
// current session. A hint is accepted when no End-User is authenticated, as the logout then only redirects to the post
// logout redirect URI and ends no session.
func oidcLogoutValidate(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, pending *model.OAuth2LogoutSession) (err error) {
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
