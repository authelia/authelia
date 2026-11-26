// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OpenIDConnectUserinfo handles GET/POST requests to the OpenID Connect 1.0 UserInfo endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
func OpenIDConnectUserinfo(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		issuer    *url.URL
		requestID uuid.UUID
		requester oauthelia2.UserinfoRequester
		err       error
	)

	if requestID, err = uuid.NewRandom(); err != nil {
		ctx.Providers.OpenIDConnect.WriteUserinfoError(ctx, rw, nil, oauthelia2.ErrServerError)

		return
	}

	ctx.GetLogger().Debugf("User Info Request with id '%s' is being processed", requestID)

	if issuer, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' could not be processed: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(rfc))

		ctx.Providers.OpenIDConnect.WriteUserinfoError(ctx, rw, nil, rfc)

		return
	}

	if requester, err = ctx.Providers.OpenIDConnect.NewUserinfoRequest(ctx, r, oidc.NewSessionWithRequestedAt(ctx.GetClock().Now())); err != nil {
		ctx.GetLogger().Errorf("User Info Request with id '%s' failed with error: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteUserinfoError(ctx, rw, requester, err)

		return
	}

	var (
		client  oidc.Client
		session *oidc.Session
		ok      bool
	)

	if client, ok = requester.GetClient().(oidc.Client); !ok {
		err = oauthelia2.ErrServerError.WithDebugf("Failed to handle client with type '%T'.", requester.GetClient())

		ctx.GetLogger().Errorf("User Info Request with id '%s' could not be processed: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteUserinfoError(ctx, rw, requester, err)

		return
	}

	ar := requester.GetAccessRequester()

	if session, ok = ar.GetSession().(*oidc.Session); !ok {
		err = oauthelia2.ErrServerError.WithDebugf("Failed to handle session with type '%T'.", ar.GetSession())

		ctx.GetLogger().Errorf("User Info Request with id '%s' on client with id '%s' could not be processed: %s", requestID, client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteUserinfoError(ctx, rw, requester, err)

		return
	}

	if !session.ValidIssuer(issuer) {
		err = oauthelia2.ErrInvalidRequest.WithDebug("The original request and the userinfo request occurred at endpoints where the origin or effective issuer did not match.")

		ctx.GetLogger().Errorf("User Info Request with id '%s' could not be processed: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteUserinfoError(ctx, rw, requester, err)

		return
	}

	claims := jwt.MapClaims{}

	if err = openIDConnectUserinfoHydrateClaims(ctx, requestID, client, ar, session, claims); err != nil {
		ctx.Providers.OpenIDConnect.WriteUserinfoError(ctx, rw, requester, err)

		return
	}

	ctx.GetLogger().Tracef("User Info Response with id '%s' on client with id '%s' is being sent with the following claims: %+v", requestID, client.GetID(), claims)

	userinfo := &oauthelia2.UserinfoClaims{}

	userinfo.FromMap(claims)

	ctx.Providers.OpenIDConnect.WriteUserinfoResponse(ctx, rw, requester, userinfo)

	ctx.GetLogger().Debugf("User Info Request with id '%s' on client with id '%s' was successfully processed", requestID, client.GetID())
}

func openIDConnectUserinfoHydrateClaims(ctx *middlewares.AutheliaCtx, requestID uuid.UUID, client oidc.Client, requester oauthelia2.AccessRequester, session *oidc.Session, claims jwt.MapClaims) (err error) {
	var detailer oidc.UserDetailer

	original := session.IDTokenClaims().ToMap()

	if session.ClientCredentials {
		if err = client.GetClaimsStrategy().HydrateClientCredentialsUserInfoClaims(ctx, client, original, claims); err != nil {
			ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' on client with id '%s' failed due to an error populating claims for the client credentials flow", requestID, client.GetID())

			return oauthelia2.ErrServerError.WithDebugf("Error occurred populating claims for the client credentials flow: %v.", err)
		}

		return nil
	}

	if detailer, err = oidc.UserDetailerFromClaims(ctx, original); err != nil {
		ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' on client with id '%s' failed due to an error loading user information", requestID, client.GetID())

		if errors.Is(err, authentication.ErrUserNotFound) {
			return oauthelia2.ErrInvalidToken.WithHint("The Access Token is missing, malformed, or invalid.").WithDebug("The user the Access Token was issued to could not be found or is disabled.")
		}

		return oauthelia2.ErrServerError.WithDebugf("Error occurred loading user information: %v.", err)
	}

	if err = client.GetClaimsStrategy().HydrateUserInfoClaims(ctx, ctx.Providers.OpenIDConnect.GetScopeStrategy(ctx), client, requester.GetGrantedScopes(), session.GrantedClaims, session.ClaimRequests.GetUserInfoRequests(), detailer, session.GetRequestedAt(), ctx.GetClock().Now(), original, claims); err != nil {
		ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' on client with id '%s' failed due to an error populating claims for the standard flow", requestID, client.GetID())

		return oauthelia2.ErrServerError.WithDebugf("Error occurred populating claims for the standard flow: %v.", err)
	}

	return nil
}
