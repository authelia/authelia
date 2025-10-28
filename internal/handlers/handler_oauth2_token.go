package handlers

import (
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"

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

	client := requester.GetClient()

	if c, ok := client.(oidc.Client); ok {
		ctx.Logger.WithFields(map[string]any{
			"client_id":                               c.GetID(),
			"access_token_signed_response_kid":        c.GetAccessTokenSignedResponseKeyID(),
			"access_token_signed_response_alg":        c.GetAccessTokenSignedResponseAlg(),
			"access_token_encrypted_response_kid":     c.GetAccessTokenEncryptedResponseKeyID(),
			"access_token_encrypted_response_alg":     c.GetAccessTokenEncryptedResponseAlg(),
			"access_token_encrypted_response_enc":     c.GetAccessTokenEncryptedResponseEnc(),
			"enable_jwt_profile_oauth2_access_tokens": c.GetEnableJWTProfileOAuthAccessTokens(),
		}).Tracef("Access Request with id '%s' is being handled by a client", requester.GetID())
	}

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' is being processed", requester.GetID(), client.GetID())

	if handled := handleOAuth2TokenHydration(ctx, rw, req, requester, responder, client, session); handled {
		return
	}

	ctx.Logger.Tracef("Access Request with id '%s' on client with id '%s' response is being generated for session with type '%T'", requester.GetID(), client.GetID(), requester.GetSession())

	if responder, err = ctx.Providers.OpenIDConnect.NewAccessResponse(ctx, requester); err != nil {
		ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %s", requester.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

		return
	}

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' has successfully been processed", requester.GetID(), client.GetID())

	ctx.Logger.Tracef("Access Request with id '%s' on client with id '%s' produced the following claims: %+v", requester.GetID(), client.GetID(), oidc.AccessResponderToClearMap(responder))

	ctx.Providers.OpenIDConnect.WriteAccessResponse(ctx, rw, requester, responder)
}

func handleOAuth2TokenHydration(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request, requester oauthelia2.AccessRequester, responder oauthelia2.AccessResponder, client oauthelia2.Client, session *oidc.Session) (handled bool) {
	var err error

	if requester.GetGrantTypes().ExactOne(oidc.GrantTypeClientCredentials) {
		if err = oidc.HydrateClientCredentialsFlowSessionWithAccessRequest(ctx, client, session); err != nil {
			ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %s", requester.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return true
		}

		if err = oidc.PopulateClientCredentialsFlowRequester(ctx, ctx.Providers.OpenIDConnect, client, requester); err != nil {
			ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %s", requester.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return true
		}
	}

	return false
}
