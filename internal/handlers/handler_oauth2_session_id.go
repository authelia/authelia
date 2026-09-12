// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func oidcSessionID(ctx *middlewares.AutheliaCtx, client oidc.Client, requester oauthelia2.Requester, userSession *session.UserSession) (sid string, err error) {
	if !requester.GetGrantedScopes().Has(oidc.ScopeOpenID) {
		return "", nil
	}

	if userSession == nil || userSession.PublicID == "" {
		return "", nil
	}

	var provider session.Strategy

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return "", err
	}

	var record *model.OAuth2SessionID

	if record, err = ctx.Providers.StorageProvider.GetOrCreateOAuth2SessionID(ctx, provider.GetIssuer(), client.GetSectorIdentifierURI(), userSession.PublicID); err != nil {
		return "", err
	}

	sid = record.SessionID.String()

	if err = ctx.Providers.StorageProvider.SaveOAuth2SessionIDClient(ctx, provider.GetIssuer(), userSession.PublicID, sid, client.GetID()); err != nil {
		return "", err
	}

	return sid, nil
}
