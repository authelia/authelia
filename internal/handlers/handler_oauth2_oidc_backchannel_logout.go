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
	"github.com/authelia/authelia/v4/internal/oidc"
)

const oidcBackChannelLogoutTimeout = time.Second * 10

type oidcBackChannelLogoutGroup struct {
	sid      string
	sectorID string
	clients  []oauthelia2.Client
}

func oidcBackChannelLogoutSession(ctx *middlewares.AutheliaCtx, issuer, username, publicID string) {
	if ctx.Providers.OpenIDConnect == nil || publicID == "" {
		return
	}

	records, err := ctx.Providers.StorageProvider.LoadOAuth2SessionIDClientsByPublicID(ctx, issuer, publicID)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s': could not load the participating clients", username)

		return
	}

	groups := oidcBackChannelLogoutGroups(ctx, username, records)

	if len(groups) == 0 {
		return
	}

	if username == "" {
		ctx.GetLogger().Warnf("Back-Channel Logout requests for the session with public identifier '%s' were skipped: the user could not be determined", publicID)

		return
	}

	timeout, cancel, err := oidcBackChannelLogoutContext(ctx)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s': could not determine the issuer", username)

		return
	}

	defer cancel()

	for _, group := range groups {
		subject, err := ctx.Providers.OpenIDConnect.GetSubject(ctx, group.sectorID, username)
		if err != nil {
			ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s': could not determine the subject identifier for the '%s' sector identifier", username, group.sectorID)

			continue
		}

		oidcBackChannelLogoutSend(ctx, timeout, username, subject.String(), group)
	}
}

func oidcBackChannelLogoutSubject(ctx *middlewares.AutheliaCtx, subject, sectorID string, clients []oidc.Client) {
	group := oidcBackChannelLogoutGroup{sectorID: sectorID}

	for _, client := range clients {
		if client.GetBackChannelLogoutURI() != "" {
			group.clients = append(group.clients, client)
		}
	}

	if len(group.clients) == 0 {
		return
	}

	timeout, cancel, err := oidcBackChannelLogoutContext(ctx)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for the subject '%s': could not determine the issuer", subject)

		return
	}

	defer cancel()

	oidcBackChannelLogoutSend(ctx, timeout, subject, subject, group)
}

// oidcBackChannelLogoutContext returns the context bounding every delivery of a single logout. The Logout Tokens are
// generated on a goroutine per Relying Party and each resolves the issuer from the context, which reads the request
// headers, so the request is detached from rather than shared with them.
func oidcBackChannelLogoutContext(ctx *middlewares.AutheliaCtx) (timeout context.Context, cancel context.CancelFunc, err error) {
	return oidc.NewDetachedContext(ctx, oidcBackChannelLogoutTimeout)
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

func oidcBackChannelLogoutSend(ctx *middlewares.AutheliaCtx, timeout context.Context, label, subject string, group oidcBackChannelLogoutGroup) {
	requester := oauthelia2.NewBackChannelLogoutRequest(subject, group.sid, group.clients)

	results, err := ctx.Providers.OpenIDConnect.SendBackChannelLogout(timeout, requester)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred delivering Back-Channel Logout requests for user '%s' with the '%s' sector identifier", label, group.sectorID)

		return
	}

	for _, result := range results {
		switch {
		case result.Skipped:
			ctx.GetLogger().Debugf("Back-Channel Logout request for user '%s' on client with id '%s' was skipped: %s", label, result.ClientID, result.Reason)
		case result.Err != nil:
			ctx.GetLogger().WithError(result.Err).Warnf("Back-Channel Logout request for user '%s' on client with id '%s' failed with status code '%d'", label, result.ClientID, result.Status)
		default:
			ctx.GetLogger().Debugf("Back-Channel Logout request for user '%s' on client with id '%s' was acknowledged with status code '%d'", label, result.ClientID, result.Status)
		}
	}
}
