package handlers

import (
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OAuth2TokenPOST handles POST requests to the OpenID Connect 1.0 Token endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
func OAuth2TokenPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		requester oauthelia2.AccessRequester
		responder oauthelia2.AccessResponder
		err       error
	)

	session := oidc.NewSessionWithRequestedAt(ctx.Clock.Now())

	if requester, err = ctx.Providers.OpenIDConnect.NewAccessRequest(ctx, req, session); err != nil {
		ctx.Logger.Errorf("Access Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

		return
	}

	client, ok := requester.GetClient().(oidc.Client)
	if !ok {
		err = oauthelia2.ErrServerError.WithDebug("The requester contained an unknown client implementation")

		ctx.Logger.Errorf("Access Response for Request with id '%s' failed with error: %s", requester.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

		return
	}

	ctx.GetLogger().
		WithFields(map[string]any{
			"access_request_id":                       requester.GetID(),
			"client_id":                               client.GetID(),
			"subject":                                 session.Subject,
			"access_token_signed_response_kid":        client.GetAccessTokenSignedResponseKeyID(),
			"access_token_signed_response_alg":        client.GetAccessTokenSignedResponseAlg(),
			"access_token_encrypted_response_kid":     client.GetAccessTokenEncryptedResponseKeyID(),
			"access_token_encrypted_response_alg":     client.GetAccessTokenEncryptedResponseAlg(),
			"access_token_encrypted_response_enc":     client.GetAccessTokenEncryptedResponseEnc(),
			"enable_jwt_profile_oauth2_access_tokens": client.GetEnableJWTProfileOAuthAccessTokens(),
		}).
		Trace("Access Request is being handled by a provider")

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' is being processed", requester.GetID(), client.GetID())

	if handled := handleOAuth2TokenHydration(ctx, rw, requester, client, session); handled {
		return
	}

	result := session.GetJWTClaims().(*jwt.JWTClaims)

	ctx.GetLogger().WithFields(map[string]any{
		"access_request_id": requester.GetID(),
		"client_id":         client.GetID(),
		"subject":           session.Subject,
		"extra":             result.Extra,
	}).Debug("Access Request Claims Result")

	ctx.GetLogger().Tracef("Access Request with id '%s' on client with id '%s' response is being generated for session with type '%T'", requester.GetID(), client.GetID(), requester.GetSession())

	if responder, err = ctx.Providers.OpenIDConnect.NewAccessResponse(ctx, requester); err != nil {
		ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %s", requester.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

		return
	}

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' has successfully been processed", requester.GetID(), client.GetID())

	ctx.Logger.Tracef("Access Request with id '%s' on client with id '%s' produced the following claims: %+v", requester.GetID(), client.GetID(), oidc.AccessResponderToClearMap(responder))

	ctx.Providers.OpenIDConnect.WriteAccessResponse(ctx, rw, requester, responder)
}

func handleOAuth2TokenHydration(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session) (handled bool) {
	var err error

	if requester.GetGrantTypes().ExactOne(oidc.GrantTypeClientCredentials) {
		if err = oidc.HydrateClientCredentialsFlowSessionWithAccessRequest(ctx, client, session); err != nil {
			ctx.GetLogger().
				WithFields(map[string]any{"oauth2_access_request_id": requester.GetID()}).
				WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
				Error("Access Request encountered an error while trying to hydrate the Client Credentials Flow claims")

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return true
		}

		if err = oidc.PopulateClientCredentialsFlowRequester(ctx, ctx.Providers.OpenIDConnect, client, requester); err != nil {
			ctx.GetLogger().
				WithFields(map[string]any{"oauth2_access_request_id": requester.GetID()}).
				WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
				Error("Access Request encountered an error while trying to populate the Client Credentials Flow requester")

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return true
		}

		return false
	}

	if client.GetEnableJWTProfileOAuthAccessTokens() {
		ctx.GetLogger().WithFields(map[string]any{"subject": session.Subject, "scope": requester.GetGrantedScopes()}).Debug("Hydrate JWT Profile Access Token claims")

		if len(session.Subject) == 0 {
			ctx.GetLogger().
				WithFields(map[string]any{"oauth2_access_request_id": requester.GetID()}).
				Trace("Access Request JWT Profile Claims Processing Skipped as no subject was provided")

			return false
		}

		var detailer oidc.UserDetailer

		if detailer, err = oidc.UserDetailerFromSubjectString(ctx, session.Subject); err != nil {
			ctx.GetLogger().
				WithFields(map[string]any{"oauth2_access_request_id": requester.GetID()}).
				WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
				Error("Access Request encountered an error while trying to obtain the detailer to hydrate the JWT Profile Access Token claims")

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return true
		}

		extra := map[string]any{}

		if err = client.GetClaimsStrategy().HydrateAccessTokenClaims(ctx, ctx.Providers.OpenIDConnect.GetScopeStrategy(ctx), client, requester.GetGrantedScopes(), nil, nil, detailer, requester.GetRequestedAt(), ctx.GetClock().Now(), nil, extra); err != nil {
			ctx.GetLogger().
				WithFields(map[string]any{"oauth2_access_request_id": requester.GetID()}).
				WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
				Error("Access Request encountered an error while trying to hydrate the JWT Profile Access Token claims")

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return true
		}

		ctx.GetLogger().WithFields(map[string]any{"extra": extra}).Debug("Access Request JWT Profile Claims Result")

		if len(extra) != 0 {
			if session.AccessToken == nil {
				session.AccessToken = &oidc.AccessTokenSession{
					Headers: map[string]any{},
					Claims:  extra,
				}
			} else {
				session.AccessToken.Claims = extra
			}
		}
	}

	return false
}
