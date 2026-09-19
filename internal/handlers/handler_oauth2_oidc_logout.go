// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

type oidcLogoutTarget struct {
	publicID string
	username string
	current  bool
}

func oidcLogout(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, pending *session.OpenIDConnectLogout) {
	if ctx.Configuration.IdentityProviders.OIDC == nil {
		return
	}

	provider, err := ctx.GetSessionProvider()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred obtaining the session provider during logout")

		return
	}

	issuer := provider.GetIssuer()

	if target := oidcLogoutResolveTarget(ctx, issuer, userSession, pending); target.publicID != "" {
		oidcLogoutSession(ctx, issuer, target)

		return
	}

	if pending != nil && pending.Subject != "" && pending.ClientID != "" && ctx.Providers.OpenIDConnect != nil {
		oidcLogoutSubject(ctx, pending.ClientID, pending.Subject)
	}
}

func oidcLogoutResolveTarget(ctx *middlewares.AutheliaCtx, issuer string, userSession *session.UserSession, pending *session.OpenIDConnectLogout) (target oidcLogoutTarget) {
	authenticated := userSession != nil && userSession.Username != "" && userSession.PublicID != ""

	if authenticated {
		target = oidcLogoutTarget{publicID: userSession.PublicID, username: userSession.Username, current: true}
	}

	if pending == nil || pending.SessionID == "" {
		return target
	}

	record, err := ctx.Providers.StorageProvider.LoadOAuth2SessionIDBySessionID(ctx, issuer, pending.SessionID)

	switch {
	case err != nil:
		ctx.GetLogger().WithError(err).Error("Error occurred loading the session identified by the 'id_token_hint' during logout")

		return target
	case record == nil:
		// The session is unknown as it has already ended or expired.
		return target
	case authenticated:
		if record.PublicID != userSession.PublicID {
			// The End-User has since logged in to another session, which is the one they confirmed the logout from.
			ctx.GetLogger().Debugf("The session identified by the 'id_token_hint' is not the session of user '%s' which is being logged out instead", userSession.Username)
		}

		return target
	default:
		return oidcLogoutTarget{publicID: record.PublicID, username: oidcLogoutUsername(ctx, pending.Subject)}
	}
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
func oidcLogoutSession(ctx *middlewares.AutheliaCtx, issuer string, target oidcLogoutTarget) {
	// Delivered before the session identifiers are removed as the participating clients are recorded against them.
	oidcBackChannelLogoutSession(ctx, issuer, target.username, target.publicID)

	records, err := ctx.Providers.StorageProvider.LoadOAuth2SessionIDsByPublicID(ctx, issuer, target.publicID)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred loading the OpenID Connect session identifiers during logout")
	}

	for _, record := range records {
		if err = ctx.Providers.StorageProvider.RevokeOAuth2SessionsBySessionID(ctx, record.SessionID.String()); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred revoking the OAuth 2.0 sessions of an OpenID Connect session during logout")
		}
	}

	if err = ctx.Providers.StorageProvider.DeleteOAuth2SessionIDByPublicID(ctx, issuer, target.publicID); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred removing the OpenID Connect session identifiers during logout")
	}

	if !target.current {
		oidcLogoutDestroySession(ctx, issuer, target)
	}
}

// oidcLogoutDestroySession ends the Authelia session identified by the 'id_token_hint', which is not the session the
// logout request was made with and so is not destroyed alongside it.
func oidcLogoutDestroySession(ctx *middlewares.AutheliaCtx, issuer string, target oidcLogoutTarget) {
	repository := ctx.Providers.SessionRepository

	if repository == nil {
		return
	}

	record, err := repository.GetByPublicID(ctx, issuer, target.publicID)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred loading the session identified by the 'id_token_hint' during logout")

		return
	}

	if record == nil {
		return
	}

	if err = repository.Delete(ctx, issuer, record.GetSessionSignature(), target.publicID, target.username); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred destroying the session identified by the 'id_token_hint' during logout")
	}
}

// oidcLogoutSubject ends every OpenID Connect session of the pairwise subject at the clients which share the sector of
// the client which requested the logout. The subject is pairwise, so it only identifies the End-User to that sector.
func oidcLogoutSubject(ctx *middlewares.AutheliaCtx, clientID, subject string) {
	client, err := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred during logout: the client with id '%s' which requested it is not registered", clientID)

		return
	}

	sectorID := client.GetSectorIdentifierURI()

	var participants []oidc.Client

	for _, c := range oidcLogoutSectorClients(ctx, sectorID) {
		has, err := ctx.Providers.StorageProvider.HasOAuth2SessionsByClientIDAndSubject(ctx, c.GetID(), subject)
		if err != nil {
			ctx.GetLogger().WithError(err).Errorf("Error occurred during logout determining if the client with id '%s' holds a session for the subject", c.GetID())

			continue
		}

		if has {
			participants = append(participants, c)
		}
	}

	oidcBackChannelLogoutSubject(ctx, subject, sectorID, participants)

	for _, c := range participants {
		if err = ctx.Providers.StorageProvider.RevokeOAuth2SessionsByClientIDAndSubject(ctx, c.GetID(), subject); err != nil {
			ctx.GetLogger().WithError(err).Errorf("Error occurred revoking the OAuth 2.0 sessions of the client with id '%s' during logout", c.GetID())
		}
	}
}

// oidcLogoutSectorClients returns the registered clients with the sector identifier.
func oidcLogoutSectorClients(ctx *middlewares.AutheliaCtx, sectorID string) (clients []oidc.Client) {
	for _, config := range ctx.Configuration.IdentityProviders.OIDC.Clients {
		client, err := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, config.ID)
		if err != nil || client.GetSectorIdentifierURI() != sectorID {
			continue
		}

		clients = append(clients, client)
	}

	return clients
}
