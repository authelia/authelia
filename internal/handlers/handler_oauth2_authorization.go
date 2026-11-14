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
	if handleOAuth2AuthorizationBounce(ctx, rw, r) {
		return
	}

	if handleOAuth2AuthorizationConsentReference(ctx, rw, r) {
		return
	}

	var (
		issuer    *url.URL
		responder oauthelia2.AuthorizeResponder
		client    oidc.Client
		policy    oidc.ClientAuthorizationPolicy
		err       error
	)

	var requester oauthelia2.AuthorizeRequester = oauthelia2.NewAuthorizeRequest()

	if issuer, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("Authorization Request could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(rfc))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, rfc)

		return
	}

	if requester, err = ctx.Providers.OpenIDConnect.NewAuthorizeRequest(ctx, r); requester == nil {
		requester = oauthelia2.NewAuthorizeRequest()

		err = oauthelia2.ErrServerError.WithDebug("The requester was nil.")

		ctx.GetLogger().Errorf("Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if requester.GetResponseMode() == oidc.ResponseModeFormPost {
		ctx.SetUserValue(middlewares.UserValueKeyOpenIDConnectResponseModeFormPost, true)
	}

	if err != nil {
		ctx.GetLogger().Errorf("Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	clientID := requester.GetClient().GetID()

	ctx.GetLogger().Debugf("Authorization Request with id '%s' on client with id '%s' is being processed", requester.GetID(), clientID)

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) || errors.Is(err, oauthelia2.ErrInvalidClient) {
			ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %s", requester.GetID(), clientID, oauthelia2.ErrorToDebugRFC6749Error(err))
		}

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	policy = client.GetAuthorizationPolicy()

	if !oidc.IsPushedAuthorizedRequest(requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)) {
		if err = client.ValidateResponseModePolicy(requester); err != nil {
			ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' failed to validate the Response Mode: %s", requester.GetID(), client.GetID(), policy.Name, oauthelia2.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

			return
		}
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		provider    session.Strategy
		handled     bool
	)

	if provider, err = ctx.GetSessionProvider(); err != nil {
		ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred obtaining session information: %+v", requester.GetID(), client.GetID(), policy.Name, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	var current *session.UserSession

	if current, err = provider.Get(ctx); err != nil {
		ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred obtaining session information: %+v", requester.GetID(), client.GetID(), policy.Name, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	userSession = *current

	var details authentication.UserDetailsExtended

	if details, err = authentication.MustGetUserDetailsExtendedSafe(userSession.Username, ctx.GetUserProvider()); err != nil {
		if !errors.Is(err, authentication.ErrUserNotFound) {
			ctx.GetLogger().WithError(err).WithField("username", userSession.Username).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred looking up user details", requester.GetID(), client.GetID(), policy.Name)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the users details."))

			return
		}

		if handled = handleOAuth2AuthorizationUserNotFound(ctx, provider, &userSession, client, policy, rw, requester); handled {
			return
		}

		details = authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
	}

	if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptNone && userSession.IsAnonymous() {
		ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: the 'prompt' type of 'none' was requested but the user is not logged in", requester.GetID(), client.GetID(), policy.Name)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrLoginRequired)

		return
	}

	if consent, handled = handleOAuth2AuthorizationConsent(ctx, issuer, client, policy, provider, userSession, &details, rw, r, requester); handled {
		return
	}

	requester.SetRequestedAt(consent.RequestedAt)

	var requests *oidc.ClaimsRequests

	extra := map[string]any{}

	if requests, handled = handleOAuth2AuthorizationClaims(ctx, rw, r, "Authorization", userSession, &details, client, requester, issuer, consent, extra); handled {
		return
	}

	ctx.GetLogger().Debugf("Authorization Request with id '%s' on client with id '%s' was successfully processed, proceeding to build Authorization Response", requester.GetID(), clientID)

	session := oidc.NewSessionWithRequester(ctx, issuer, ctx.Providers.OpenIDConnect.Issuer.GetKeyID(ctx, client.GetIDTokenSignedResponseKeyID(), client.GetIDTokenSignedResponseAlg()), details.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extra, userSession.LastAuthenticatedTime(), consent, requester, requests)

	if client.GetClaimsStrategy().MergeAccessTokenAudienceWithIDTokenAudience() {
		session.Claims.Audience = append([]string{clientID}, oauthelia2.JoinGrantedAudienceAndResource(requester.GetGrantedAudience(), requester.GetGrantedResource())...)
	}

	ctx.GetLogger().Tracef("Authorization Request with id '%s' on client with id '%s' using policy '%s' creating session for Authorization Response for subject '%s' with username '%s' with groups: %+v and claims: %+v",
		requester.GetID(), session.ClientID, policy.Name, session.Subject, session.Username, details.Groups, session.Claims)

	ctx.GetLogger().WithFields(map[string]any{"id": requester.GetID(), "response_type": requester.GetResponseTypes(), "response_mode": requester.GetResponseMode(), "scope": requester.GetRequestedScopes(), "aud": requester.GetRequestedAudience(), "resource": requester.GetRequestedResource(), "redirect_uri": requester.GetRedirectURI(), "state": requester.GetState()}).Tracef("Authorization Request is using the following request parameters")

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(ctx, consent.ID); err != nil {
		ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), policy.Name, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return
	}

	if responder, err = ctx.Providers.OpenIDConnect.NewAuthorizeResponse(ctx, requester, session); err != nil {
		ctx.GetLogger().Errorf("Authorization Response for Request with id '%s' on client with id '%s' using policy '%s' could not be created: %s", requester.GetID(), clientID, policy.Name, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	ctx.Providers.OpenIDConnect.WriteAuthorizeResponse(ctx, rw, requester, responder)
}

func handleOAuth2AuthorizationUserNotFound(ctx *middlewares.AutheliaCtx, provider session.Strategy, userSession *session.UserSession, client oidc.Client, policy oidc.ClientAuthorizationPolicy, rw http.ResponseWriter, requester oauthelia2.AuthorizeRequester) (handled bool) {
	var err error

	ctx.GetLogger().WithField("username", userSession.Username).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not use the existing session: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login", requester.GetID(), client.GetID(), policy.Name)

	if err = provider.Destroy(ctx); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' for user '%s' had an error while destroying session", requester.GetID(), client.GetID(), policy.Name, userSession.Username)
	}

	*userSession = provider.NewDefault()

	if err = provider.Save(ctx, userSession); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' had an error while saving updated session", requester.GetID(), client.GetID(), policy.Name)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

		return true
	}

	return false
}

// OAuth2AuthorizationPOST handles Authorization Requests which use the HTTP POST method per OpenID Connect 1.0 Core
// Section 3.1.2.1. The session cookie is not sent with a cross-site POST so the request is validated, stored in a consent
// session bound to the user agent, and the user agent is redirected to the Authorization Endpoint with a reference to
// it, which the GET handler resolves. This keeps the parameters out of the URL. The redirection uses the 303 status
// code as it's the only status code which unambiguously instructs the user-agent to rewrite the request method to GET,
// and per the FAPI 2.0 Security Profile Section 5.3.2.2 the authorization server should use 303 when redirecting the
// user-agent.
//
// https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint
//
// https://openid.net/specs/fapi-security-profile-2_0-final.html
func OAuth2AuthorizationPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		issuer    *url.URL
		client    oidc.Client
		consent   *model.OAuth2ConsentSession
		requester oauthelia2.AuthorizeRequester
		err       error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("Authorization Request could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(rfc))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, oauthelia2.NewAuthorizeRequest(), rfc)

		return
	}

	r.Body = http.MaxBytesReader(rw, r.Body, 10<<20)

	if requester, err = ctx.Providers.OpenIDConnect.NewAuthorizeRequest(ctx, r); requester == nil {
		requester = oauthelia2.NewAuthorizeRequest()

		err = oauthelia2.ErrServerError.WithDebug("The requester was nil.")

		ctx.GetLogger().Errorf("Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if requester.GetResponseMode() == oidc.ResponseModeFormPost {
		ctx.SetUserValue(middlewares.UserValueKeyOpenIDConnectResponseModeFormPost, true)
	}

	if err != nil {
		ctx.GetLogger().Errorf("Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	clientID := requester.GetClient().GetID()

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %s", requester.GetID(), clientID, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	prefixPAR := ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)

	if !oidc.IsPushedAuthorizedRequest(requester, prefixPAR) {
		if err = client.ValidateResponseModePolicy(requester); err != nil {
			ctx.GetLogger().Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' failed to validate the Response Mode: %s", requester.GetID(), clientID, client.GetAuthorizationPolicy().Name, oauthelia2.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

			return
		}
	}

	if consent, err = handleOAuth2NewConsentSession(ctx, uuid.Nil, requester, prefixPAR); err != nil {
		ctx.GetLogger().Errorf(logFmtErrConsentGenerateError, requester.GetID(), clientID, client.GetConsentPolicy(), "generating", err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
		ctx.GetLogger().Errorf(logFmtErrConsentGenerateError, requester.GetID(), clientID, client.GetConsentPolicy(), "saving", err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return
	}

	handleOAuth2PushedAuthorizeConsent(ctx, requester, requester.GetRequestForm())

	ctx.GetLogger().Debugf("Authorization Request with id '%s' on client with id '%s' was stored in the consent session with id '%s' and is being redirected to the Authorization Endpoint", requester.GetID(), clientID, consent.ChallengeID)

	http.Redirect(rw, r, handleOAuth2AuthorizationConsentReferenceURL(issuer, consent).String(), http.StatusSeeOther)
}
