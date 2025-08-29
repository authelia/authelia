package handlers

import (
	"errors"
	"net/http"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OAuth2AuthorizationGET handles GET/POST requests to the OpenID Connect 1.0 Authorization endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint
//
//nolint:gocyclo
func OAuth2AuthorizationGET(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.AuthorizeRequester
		responder oauthelia2.AuthorizeResponder
		client    oidc.Client
		policy    oidc.ClientAuthorizationPolicy
		err       error
	)
	if requester, err = ctx.Providers.OpenIDConnect.NewAuthorizeRequest(ctx, r); requester == nil {
		err = oauthelia2.ErrServerError.WithDebug("The requester was nil.")

		ctx.Logger.Errorf("Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if requester.GetResponseMode() == oidc.ResponseModeFormPost {
		ctx.SetUserValue(middlewares.UserValueKeyOpenIDConnectResponseModeFormPost, true)
	}

	if err != nil {
		ctx.Logger.Errorf("Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	clientID := requester.GetClient().GetID()

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being processed", requester.GetID(), clientID)

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %s", requester.GetID(), clientID, oauthelia2.ErrorToDebugRFC6749Error(err))
		}

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	policy = client.GetAuthorizationPolicy()

	if !oidc.IsPushedAuthorizedRequest(requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)) {
		if err = client.ValidateResponseModePolicy(requester); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' failed to validate the Response Mode: %s", requester.GetID(), client.GetID(), policy.Name, oauthelia2.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

			return
		}
	}

	var (
		issuer      *url.URL
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		provider    *session.Session
		handled     bool
	)

	if provider, err = ctx.GetSessionProvider(); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred obtaining session information: %+v", requester.GetID(), client.GetID(), policy.Name, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred obtaining session information: %+v", requester.GetID(), client.GetID(), policy.Name, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptNone && userSession.IsAnonymous() {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: the 'prompt' type of 'none' was requested but the user is not logged in", requester.GetID(), client.GetID(), policy.Name)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrLoginRequired)

		return
	}

	issuer = ctx.RootURL()

	if consent, handled = handleOAuth2AuthorizationConsent(ctx, issuer, client, policy, provider, userSession, rw, r, requester); handled {
		return
	}

	requester.SetRequestedAt(consent.RequestedAt)

	var details *authentication.UserDetailsExtended

	if details, err = ctx.Providers.UserProvider.GetDetailsExtended(userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred retrieving user details for '%s' from the backend", requester.GetID(), client.GetID(), policy.Name, userSession.Username)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the users details."))

		return
	}

	var requests *oidc.ClaimsRequests

	extra := map[string]any{}

	if requests, handled = handleOAuth2AuthorizationClaims(ctx, rw, r, "Authorization", userSession, details, client, requester, issuer, consent, extra); handled {
		return
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' was successfully processed, proceeding to build Authorization Response", requester.GetID(), clientID)

	session := oidc.NewSessionWithRequester(ctx, issuer, ctx.Providers.OpenIDConnect.Issuer.GetKeyID(ctx, client.GetIDTokenSignedResponseKeyID(), client.GetIDTokenSignedResponseAlg()), details.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extra, userSession.LastAuthenticatedTime(), consent, requester, requests)

	if client.GetClaimsStrategy().MergeAccessTokenAudienceWithIDTokenAudience() {
		session.Claims.Audience = append([]string{clientID}, requester.GetGrantedAudience()...)
	}

	ctx.Logger.Tracef("Authorization Request with id '%s' on client with id '%s' using policy '%s' creating session for Authorization Response for subject '%s' with username '%s' with groups: %+v and claims: %+v",
		requester.GetID(), session.ClientID, policy.Name, session.Subject, session.Username, userSession.Groups, session.Claims)

	ctx.Logger.WithFields(map[string]any{"id": requester.GetID(), "response_type": requester.GetResponseTypes(), "response_mode": requester.GetResponseMode(), "scope": requester.GetRequestedScopes(), "aud": requester.GetRequestedAudience(), "redirect_uri": requester.GetRedirectURI(), "state": requester.GetState()}).Tracef("Authorization Request is using the following request parameters")

	if responder, err = ctx.Providers.OpenIDConnect.NewAuthorizeResponse(ctx, requester, session); err != nil {
		ctx.Logger.Errorf("Authorization Response for Request with id '%s' on client with id '%s' using policy '%s' could not be created: %s", requester.GetID(), clientID, policy.Name, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(ctx, consent.ID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), policy.Name, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return
	}

	ctx.Providers.OpenIDConnect.WriteAuthorizeResponse(ctx, rw, requester, responder)
}

// OAuth2AuthorizationPOST handles redirecting users to use the GET request to ensure the session cookie is
// included if available.
func OAuth2AuthorizationPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var err error
	if err = r.ParseMultipartForm(1 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		requester := oauthelia2.NewAuthorizeRequest()

		ctx.Logger.WithError(err).Errorf("Authorization Request with id '%s' had an error parsing a multipart form.", requester.GetID())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	query := r.Form

	redirectURL := ctx.RootURL()

	redirectURL = redirectURL.JoinPath(oidc.EndpointPathAuthorization)

	redirectURL.RawQuery = query.Encode()

	http.Redirect(rw, r, redirectURL.String(), http.StatusFound)
}

// OAuth2PushedAuthorizationRequest handles POST requests to the OAuth 2.0 Pushed Authorization Requests endpoint.
//
// RFC9126 https://www.rfc-editor.org/rfc/rfc9126.html
func OAuth2PushedAuthorizationRequest(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.AuthorizeRequester
		responder oauthelia2.PushedAuthorizeResponder
		err       error
	)
	if requester, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeRequest(ctx, r); err != nil {
		ctx.Logger.Errorf("Pushed Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	var client oidc.Client

	clientID := requester.GetClient().GetID()

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) {
			ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %+v", requester.GetID(), clientID, err)
		}

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = client.ValidateResponseModePolicy(requester); err != nil {
		ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' failed to validate the Response Modes: %s", requester.GetID(), client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	if responder, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeResponse(ctx, requester, oidc.NewSessionWithRequestedAt(ctx.GetClock().Now())); err != nil {
		ctx.Logger.Errorf("Pushed Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	ctx.Providers.OpenIDConnect.WritePushedAuthorizeResponse(ctx, rw, requester, responder)
}
