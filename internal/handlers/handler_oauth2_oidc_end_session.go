package handlers

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OpenIDConnectEndSession handles requests made by resource owners when the relying-party redirects them
// requesting they logout.
//
// OpenID Connect RP-Initiated Logout 1.0 (https://openid.net/specs/openid-connect-rpinitiated-1_0.html)
func OpenIDConnectEndSession(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		issuer *url.URL
		err    error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred determining issuer")

		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	var (
		tokenString = req.PostFormValue(oidc.FormParameterIDTokenHint)
		id          = req.PostFormValue(oidc.FormParameterClientID)
		redirect    = req.PostFormValue(oidc.FormParameterPostLogoutRedirectURI)
		state       = req.PostFormValue(oidc.FormParameterState)
	)

	if len(tokenString) == 0 {
		tokenString = req.URL.Query().Get(oidc.FormParameterIDTokenHint)
	}

	if len(id) == 0 {
		id = req.URL.Query().Get(oidc.FormParameterClientID)
	}

	if len(redirect) == 0 {
		redirect = req.URL.Query().Get(oidc.FormParameterPostLogoutRedirectURI)
	}

	if len(state) == 0 {
		state = req.URL.Query().Get(oidc.FormParameterState)
	}

	var client oidc.Client

	if len(id) > 0 {
		if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, id); err != nil {
			oidcEndSessionRedirectError(ctx, rw, req, issuer,
				oauthelia2.ErrInvalidClient.
					WithHintf("The client_id '%s' is not registered with this authorization server.", id).
					WithDebugError(err))

			return
		}
	}

	if len(tokenString) > 0 {
		var athash string

		if client, athash, err = oidcEndSessionValidateIDTokenHint(ctx, issuer, client, id, tokenString); err != nil {
			oidcEndSessionRedirectError(ctx, rw, req, issuer, err)

			return
		}

		id = client.GetID()

		_ = ctx.Providers.OpenIDConnect.DeleteAccessTokenSession(ctx, athash)
	}

	if len(redirect) > 0 {
		if client == nil {
			oidcEndSessionRedirectError(ctx, rw, req, issuer,
				oauthelia2.ErrInvalidRequest.
					WithHint("The post_logout_redirect_uri parameter requires either the id_token_hint or client_id parameter to identify the client."))

			return
		}

		if !slices.Contains(client.GetPostLogoutRedirectURIs(), redirect) {
			oidcEndSessionRedirectError(ctx, rw, req, issuer,
				oauthelia2.ErrInvalidRequest.
					WithHintf("The post_logout_redirect_uri '%s' is not registered for client '%s'.", redirect, id))

			return
		}
	}

	oidcEndSessionRedirectSuccess(rw, req, issuer, redirect, state)
}

func oidcEndSessionValidateIDTokenHint(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client, id, tokenString string) (resolved oidc.Client, athash string, err error) {
	var token *jwt.Token

	if token, tokenString, err = oidc.DecodeIDTokenUnverified(ctx, ctx.Providers.OpenIDConnect.Strategy.JWT, client, tokenString); err != nil {
		return nil, "", oauthelia2.ErrInvalidRequest.
			WithHint("The id_token_hint could not be decoded.").
			WithDebugError(err)
	}

	claims, ok := token.Claims.(*jwt.IDTokenClaims)
	if !ok {
		return nil, "", oauthelia2.ErrInvalidRequest.
			WithHint("The id_token_hint claims were of an unexpected type.")
	}

	opts := []jwt.ClaimValidationOption{
		jwt.ValidateIssuer(issuer.String()),
	}

	if len(id) > 0 {
		opts = append(opts, jwt.ValidateAuthorizedParty(id), jwt.ValidateAudienceAll(id))
	}

	if err = claims.Valid(opts...); err != nil && !oidc.IsExpiredValidationError(err) {
		return nil, "", oauthelia2.ErrInvalidRequest.
			WithHint("The id_token_hint failed validation.").
			WithDebugError(err)
	}

	if client != nil {
		return client, claims.AccessTokenHash, nil
	}

	if len(claims.Audience) < 1 {
		return nil, "", oauthelia2.ErrInvalidRequest.
			WithHint("The id_token_hint did not contain any audience to identify the client.")
	}

	id = claims.Audience[0]

	if resolved, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, id); err != nil {
		return nil, "", oauthelia2.ErrInvalidClient.
			WithHintf("The client_id '%s' from the id_token_hint is not registered with this authorization server.", id).
			WithDebugError(err)
	}

	if token, err = ctx.Providers.OpenIDConnect.Strategy.JWT.Decode(ctx, tokenString, jwt.WithClient(jwt.NewIDTokenClient(resolved))); err != nil && !oidc.IsExpiredValidationError(err) {
		return nil, "", oauthelia2.ErrInvalidRequest.
			WithHint("The id_token_hint could not be decoded.").
			WithDebugError(err)
	}

	if err = token.Valid(jwt.ValidateTypes("JWT")); err != nil {
		return nil, "", oauthelia2.ErrInvalidRequest.
			WithHint("The id_token_hint failed validation.").
			WithDebugError(err)
	}

	return resolved, claims.AccessTokenHash, nil
}

func oidcEndSessionRedirectSuccess(rw http.ResponseWriter, req *http.Request, issuer *url.URL, redirect, state string) {
	location := issuer.JoinPath(oidcFrontendEndpointPathLogout)

	query := location.Query()

	query.Set(oidcFrontendQueryParameterConfirm, "true")

	if len(redirect) > 0 {
		query.Set(oidcFrontendQueryParameterRedirectionURL, redirect)

		if len(state) > 0 {
			query.Set(oidc.FormParameterState, state)
		}
	}

	location.RawQuery = query.Encode()

	rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
	rw.Header().Set(fasthttp.HeaderPragma, "no-cache")

	http.Redirect(rw, req, location.String(), http.StatusFound)
}

func oidcEndSessionRedirectError(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request, issuer *url.URL, err error) {
	rfc := oauthelia2.ErrorToRFC6749Error(err)

	ctx.Logger.WithError(err).Errorf("RP-Initiated Logout request could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

	location := issuer.JoinPath(oidc.FrontendEndpointPathConsentCompletion)

	query := location.Query()

	if len(rfc.ErrorField) != 0 {
		query.Set(oidcFrontendQueryParameterError, rfc.ErrorField)
	}

	if len(rfc.DescriptionField) != 0 {
		query.Set(oidcFrontendQueryParameterErrorDescription, rfc.DescriptionField)
	}

	if len(rfc.HintField) != 0 {
		query.Set(oidcFrontendQueryParameterErrorHint, rfc.HintField)
	}

	if ctx.Providers.OpenIDConnect != nil && ctx.Providers.OpenIDConnect.GetSendDebugMessagesToClients(ctx) && len(rfc.DebugField) != 0 {
		query.Set(oidcFrontendQueryParameterErrorDebug, rfc.DebugField)
	}

	location.RawQuery = query.Encode()

	rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
	rw.Header().Set(fasthttp.HeaderPragma, "no-cache")

	http.Redirect(rw, req, location.String(), http.StatusFound)
}

const (
	oidcFrontendEndpointPathLogout = "/logout"

	oidcFrontendQueryParameterConfirm        = "confirm"
	oidcFrontendQueryParameterRedirectionURL = "rd"

	oidcFrontendQueryParameterError            = "error"
	oidcFrontendQueryParameterErrorDescription = "error_description"
	oidcFrontendQueryParameterErrorHint        = "error_hint"
	oidcFrontendQueryParameterErrorDebug       = "error_debug"
)
