// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"

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

	return record.SessionID.String(), nil
}

var errOIDCSessionEnded = errors.New("the session the session identifier was issued for has ended")

func oidcSessionIDConfirm(ctx *middlewares.AutheliaCtx, sid string, userSession *session.UserSession) (err error) {
	if sid == "" {
		return nil
	}

	var provider session.Strategy

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return err
	}

	var exists bool

	if exists, err = provider.Exists(ctx, userSession.PublicID); err != nil {
		return err
	}

	if !exists {
		return errOIDCSessionEnded
	}

	return nil
}
