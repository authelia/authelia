// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

const oidcBackChannelLogoutTimeout = time.Second * 10

type oidcBackChannelLogoutGroup struct {
	sid      string
	sectorID string
	clients  []oauthelia2.Client
}

func oidcBackChannelLogout(ctx *middlewares.AutheliaCtx, userSession *session.UserSession) {
	if ctx.Providers.OpenIDConnect == nil || ctx.Configuration.IdentityProviders.OIDC == nil || userSession == nil || userSession.Username == "" || userSession.PublicID == "" {
		return
	}

	provider, err := ctx.GetSessionProvider()
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s': could not obtain the session provider", userSession.Username)

		return
	}

	var records []model.OAuth2SessionIDClient

	if records, err = ctx.Providers.StorageProvider.LoadOAuth2SessionIDClientsByPublicID(ctx, provider.GetIssuer(), userSession.PublicID); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s': could not load the participating clients", userSession.Username)

		return
	}

	groups := oidcBackChannelLogoutGroups(ctx, userSession.Username, records)

	if len(groups) == 0 {
		return
	}

	timeout, cancel := context.WithTimeout(context.WithoutCancel(ctx), oidcBackChannelLogoutTimeout)

	defer cancel()

	for _, group := range groups {
		oidcBackChannelLogoutSend(ctx, timeout, userSession.Username, group)
	}
}

func oidcBackChannelLogoutGroups(ctx *middlewares.AutheliaCtx, username string, records []model.OAuth2SessionIDClient) (groups []oidcBackChannelLogoutGroup) {
	index := make(map[[2]string]int)

	for _, record := range records {
		client, err := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, record.ClientID)
		if err != nil {
			ctx.GetLogger().WithError(err).Debugf("Back-Channel Logout request for user '%s' on client with id '%s' was skipped: the client is not registered", username, record.ClientID)

			continue
		}

		if client.GetBackChannelLogoutURI() == "" {
			continue
		}

		key := [2]string{record.SessionID.String(), client.GetSectorIdentifierURI()}

		i, ok := index[key]
		if !ok {
			i = len(groups)
			index[key] = i

			groups = append(groups, oidcBackChannelLogoutGroup{sid: key[0], sectorID: key[1]})
		}

		groups[i].clients = append(groups[i].clients, client)
	}

	return groups
}

func oidcBackChannelLogoutSend(ctx *middlewares.AutheliaCtx, timeout context.Context, username string, group oidcBackChannelLogoutGroup) {
	subject, err := ctx.Providers.OpenIDConnect.GetSubject(ctx, group.sectorID, username)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s': could not determine the subject identifier for the '%s' sector identifier", username, group.sectorID)

		return
	}

	requester := oauthelia2.NewBackChannelLogoutRequest(subject.String(), group.sid, group.clients)

	var results []oauthelia2.BackChannelLogoutResult

	if results, err = ctx.Providers.OpenIDConnect.SendBackChannelLogout(timeout, requester); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s' with the '%s' sector identifier", username, group.sectorID)

		return
	}

	for _, result := range results {
		switch {
		case result.Skipped:
			ctx.GetLogger().Debugf("Back-Channel Logout request for user '%s' on client with id '%s' was skipped: %s", username, result.ClientID, result.Reason)
		case result.Err != nil:
			ctx.GetLogger().WithError(result.Err).Warnf("Back-Channel Logout request for user '%s' on client with id '%s' failed with status code '%d'", username, result.ClientID, result.Status)
		default:
			ctx.GetLogger().Debugf("Back-Channel Logout request for user '%s' on client with id '%s' was acknowledged with status code '%d'", username, result.ClientID, result.Status)
		}
	}
}
