// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func oidcSessionID(ctx *middlewares.AutheliaCtx, client oidc.Client, userSession *session.UserSession) (sid string, err error) {
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

func oidcSessionIDDeviceError(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, requester oauthelia2.DeviceAuthorizeRequester, log *logrus.Entry, err error, action, hint string) {
	if errors.Is(err, errOIDCSessionEnded) {
		log.Error("Device Authorization Request failed as the user session ended while the request was being processed during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied.WithHint(messageOIDCSessionEnded))

		return
	}

	log.WithError(err).Errorf("Device Authorization Request failed to %s during the User Authorization Flow", action)

	ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint(hint))
}

func oidcSessionIDAuthorizeError(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, requester oauthelia2.AuthorizeRequester, err error, prefix, action, hint string) {
	if errors.Is(err, errOIDCSessionEnded) {
		ctx.GetLogger().Errorf("%s: the user session ended while the request was being processed", prefix)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied.WithHint(messageOIDCSessionEnded))

		return
	}

	ctx.GetLogger().WithError(err).Errorf("%s: error occurred %s", prefix, action)

	ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint(hint))
}

const messageOIDCSessionEnded = "The user session ended while the request was being processed."
