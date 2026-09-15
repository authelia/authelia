// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

const oidcBackChannelLogoutTimeout = time.Second * 10

func oidcBackChannelLogout(ctx *middlewares.AutheliaCtx, username, sid string) {
	if ctx.Providers.OpenIDConnect == nil {
		return
	}

	clients := oidcBackChannelLogoutClients(ctx, username, sid)

	if len(clients) == 0 {
		return
	}

	timeout, cancel := context.WithTimeout(ctx, oidcBackChannelLogoutTimeout)

	defer cancel()

	for sectorID, sectorClients := range oidcBackChannelLogoutClientsBySector(clients) {
		oidcBackChannelLogoutSector(ctx, timeout, username, sid, sectorID, sectorClients)
	}
}

func oidcBackChannelLogoutSector(ctx *middlewares.AutheliaCtx, timeout context.Context, username, sid, sectorID string, clients []oauthelia2.Client) {
	subject, err := ctx.Providers.OpenIDConnect.GetSubject(ctx, sectorID, username)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s': could not determine the subject identifier for the '%s' sector identifier", username, sectorID)

		return
	}

	requester := oauthelia2.NewBackChannelLogoutRequest(subject.String(), sid, clients)

	var results []oauthelia2.BackChannelLogoutResult

	if results, err = ctx.Providers.OpenIDConnect.SendBackChannelLogout(timeout, requester); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s' with the '%s' sector identifier", username, sectorID)

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

func oidcBackChannelLogoutClientsBySector(clients []oauthelia2.Client) (sectors map[string][]oauthelia2.Client) {
	sectors = make(map[string][]oauthelia2.Client)

	for _, client := range clients {
		sectorID := ""

		if c, ok := client.(oidc.Client); ok {
			sectorID = c.GetSectorIdentifierURI()
		}

		sectors[sectorID] = append(sectors[sectorID], client)
	}

	return sectors
}

// oidcBackChannelLogoutClients returns the clients which participated in the End-User session identified by the
// given username and session identifier, and which therefore must be sent a Logout Token.
//
// TODO: This provider does not record which clients participated in which End-User session, so no client is ever
// returned and no Logout Token is ever delivered. The session rewrite introduces the session identifier ('sid')
// this needs, at which point the participating clients are looked up by it and resolved through
// ctx.Providers.OpenIDConnect.GetRegisteredClient. Until then both parameters are unused.
func oidcBackChannelLogoutClients(ctx *middlewares.AutheliaCtx, username, sid string) (clients []oauthelia2.Client) {
	return nil
}
