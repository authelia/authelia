package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func handleOAuth2AuthorizationConsent(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client, policy oidc.ClientAuthorizationPolicy,
	provider *session.Session, userSession session.UserSession,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		subject uuid.UUID
		err     error
	)

	var handler handlerAuthorizationConsent

	if handled = handleOAuth2AuthorizationConsentSessionUpdates(ctx, provider, &userSession, client, policy, rw, requester); handled {
		return nil, handled
	}

	level := policy.GetRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()})

	switch {
	case userSession.IsAnonymous():
		handler = handleOAuth2AuthorizationConsentNotAuthenticated
	case authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA), level):
		if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentCantGetSubject, requester.GetID(), client.GetID(), client.GetConsentPolicy(), userSession.Username, client.GetSectorIdentifierURI(), err)

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrSubjectCouldNotLookup)

			return nil, true
		}

		switch client.GetConsentPolicy().Mode {
		case oidc.ClientConsentModeExplicit:
			handler = handleOAuth2AuthorizationConsentModeExplicit
		case oidc.ClientConsentModeImplicit:
			if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptConsent {
				handler = handleOAuth2AuthorizationConsentModeExplicit

				break
			}

			if requester.GetRequestedScopes().Has(oidc.ScopeOfflineAccess) || requester.GetRequestedScopes().Has(oidc.ScopeOffline) {
				if ar, ok := requester.(oauthelia2.AuthorizeRequester); ok && ar.GetResponseTypes().Has(oidc.ResponseTypeAuthorizationCodeFlow) {
					handler = handleOAuth2AuthorizationConsentModeExplicit

					break
				}
			}

			handler = handleOAuth2AuthorizationConsentModeImplicit
		case oidc.ClientConsentModePreConfigured:
			handler = handleOAuth2AuthorizationConsentModePreConfigured
		default:
			ctx.Logger.Errorf(logFmtErrConsentCantDetermineConsentMode, requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not determine the client consent mode."))

			return nil, true
		}
	case level == authorization.Denied:
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: the user '%s' is not authorized to use this client", requester.GetID(), client.GetID(), policy.Name, userSession.Username)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

		return nil, true
	default:
		return handleOAuth2AuthorizationConsentGenerate(ctx, issuer, client, userSession, uuid.Nil, rw, r, requester)
	}

	return handler(ctx, issuer, client, userSession, subject, rw, r, requester)
}

func handleOAuth2AuthorizationConsentSessionUpdates(ctx *middlewares.AutheliaCtx, provider *session.Session, userSession *session.UserSession, client oidc.Client, policy oidc.ClientAuthorizationPolicy, rw http.ResponseWriter, requester oauthelia2.Requester) (handled bool) {
	var err error

	if modified, invalid := handleSessionValidateRefresh(ctx, userSession, ctx.Configuration.AuthenticationBackend.RefreshInterval); invalid {
		if err = ctx.DestroySession(); err != nil {
			ctx.Logger.WithError(err).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' for user '%s' had an error while destroying session", requester.GetID(), client.GetID(), policy.Name, userSession.Username)
		}

		*userSession = provider.NewDefaultUserSession()
		userSession.LastActivity = ctx.GetClock().Now().Unix()

		if err = provider.SaveSession(ctx.RequestCtx, *userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' for user '%s' had an error while saving updated session", requester.GetID(), client.GetID(), policy.Name, userSession.Username)

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

			return true
		}
	} else if modified {
		if err = provider.SaveSession(ctx.RequestCtx, *userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' for user '%s' had an error while saving updated session", requester.GetID(), client.GetID(), policy.Name, userSession.Username)

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

			return true
		}
	}

	return false
}

func handleOAuth2AuthorizationConsentNotAuthenticated(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	_ session.UserSession, _ uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var err error
	if consent, err = handleOAuth2NewConsentSession(ctx, uuid.UUID{}, requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "saving", err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	redirectionURL := handleOIDCAuthorizationConsentGetRedirectionURL(ctx, issuer, consent)

	handleOAuth2PushedAuthorizeConsent(ctx, requester, r.Form)

	http.Redirect(rw, r, redirectionURL.String(), http.StatusFound)

	return nil, true
}

func handleOAuth2AuthorizationConsentGenerate(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	ctx.Logger.Debugf(logFmtDbgConsentGenerate, requester.GetID(), client.GetID(), client.GetConsentPolicy())

	if len(ctx.QueryArgs().PeekBytes(qryArgConsentID)) != 0 {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", errors.New("consent id value was present when it should be absent"))

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if consent, err = handleOAuth2NewConsentSession(ctx, subject, requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "saving", err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	if oidc.RequesterRequiresLogin(requester, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		handleOAuth2AuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	} else {
		ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime(), "prompt": requester.GetRequestForm().Get("prompt")}).Debugf("Authorization Request with id '%s' on client with id '%s' is not being redirected for reauthentication", requester.GetID(), client.GetID())
	}

	handleOAuth2AuthorizationConsentRedirect(ctx, issuer, consent, client, userSession, rw, r, requester)

	return consent, true
}

func handleOAuth2AuthorizationConsentRedirect(ctx *middlewares.AutheliaCtx, issuer *url.URL, consent *model.OAuth2ConsentSession, client oidc.Client,
	userSession session.UserSession, rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) {
	var location *url.URL

	if client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA), authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		location, _ = url.ParseRequestURI(issuer.String())
		location.Path = path.Join(location.Path, oidc.FrontendEndpointPathConsentDecision)

		query := location.Query()
		query.Set(queryArgFlow, flowNameOpenIDConnect)
		query.Set(queryArgFlowID, consent.ChallengeID.String())

		location.RawQuery = query.Encode()

		ctx.Logger.Debugf(logFmtDbgConsentAuthenticationSufficiency, requester.GetID(), client.GetID(), client.GetConsentPolicy(), userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA).String(), "sufficient", client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}))
	} else {
		location = handleOIDCAuthorizationConsentGetRedirectionURL(ctx, issuer, consent)

		ctx.Logger.Debugf(logFmtDbgConsentAuthenticationSufficiency, requester.GetID(), client.GetID(), client.GetConsentPolicy(), userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA).String(), "insufficient", client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}))
	}

	handleOAuth2PushedAuthorizeConsent(ctx, requester, r.Form)

	ctx.Logger.Debugf(logFmtDbgConsentRedirect, requester.GetID(), client.GetID(), client.GetConsentPolicy(), location)

	http.Redirect(rw, r, location.String(), http.StatusFound)
}

func handleOAuth2PushedAuthorizeConsent(ctx *middlewares.AutheliaCtx, requester oauthelia2.Requester, form url.Values) {
	if !oidc.IsPushedAuthorizedRequest(requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)) {
		return
	}

	par, err := ctx.Providers.StorageProvider.LoadOAuth2PARContext(ctx, form.Get(oidc.FormParameterRequestURI))
	if err != nil {
		ctx.Logger.WithError(err).Warnf("Authorization Request with id '%s' on client with id '%s' encountered a storage error while trying to make the Pushed Authorize Request session available for consent", requester.GetID(), requester.GetClient().GetID())

		return
	}

	par.Revoked = false

	if err = ctx.Providers.StorageProvider.UpdateOAuth2PARContext(ctx, *par); err != nil {
		ctx.Logger.WithError(err).Warnf("Authorization Request with id '%s' on client with id '%s' encountered a storage error while trying to make the Pushed Authorize Request session available for consent", requester.GetID(), requester.GetClient().GetID())

		return
	}
}

func handleOAuth2AuthorizationConsentPromptLoginRedirect(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client, userSession session.UserSession, rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester, consent *model.OAuth2ConsentSession) {
	ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime()}).Debugf("Authorization Request with id '%s' on client with id '%s' is being redirected for reauthentication: prompt type login was requested", requester.GetID(), client.GetID())

	handleOAuth2PushedAuthorizeConsent(ctx, requester, r.Form)

	redirectionURL := issuer.JoinPath(oidc.FrontendEndpointPathConsentDecision)

	query := redirectionURL.Query()
	query.Set(queryArgFlow, flowNameOpenIDConnect)
	query.Set(queryArgFlowID, consent.ChallengeID.String())

	redirectionURL.RawQuery = query.Encode()

	http.Redirect(rw, r, redirectionURL.String(), http.StatusFound)
}

func handleOIDCAuthorizationConsentGetRedirectionURL(_ *middlewares.AutheliaCtx, issuer *url.URL, consent *model.OAuth2ConsentSession) (redirectURL *url.URL) {
	iss := issuer.String()

	if !strings.HasSuffix(iss, "/") {
		iss += "/"
	}

	redirectURL, _ = url.ParseRequestURI(iss)

	query := redirectURL.Query()
	query.Set(queryArgFlow, flowNameOpenIDConnect)
	query.Set(queryArgFlowID, consent.ChallengeID.String())

	redirectURL.RawQuery = query.Encode()

	return redirectURL
}

func handleOAuth2NewConsentSession(ctx oidc.Context, subject uuid.UUID, requester oauthelia2.Requester, prefixPAR string) (consent *model.OAuth2ConsentSession, err error) {
	if oidc.IsPushedAuthorizedRequest(requester, prefixPAR) {
		form := url.Values{}

		form.Set(oidc.FormParameterRequestURI, requester.GetRequestForm().Get(oidc.FormParameterRequestURI))

		if requester.GetRequestForm().Has(oidc.FormParameterClientID) {
			form.Set(oidc.FormParameterClientID, requester.GetRequestForm().Get(oidc.FormParameterClientID))
		}

		return model.NewOAuth2ConsentSessionWithForm(ctx.GetClock().Now().UTC().Add(time.Minute*10), subject, requester, form)
	}

	return model.NewOAuth2ConsentSession(ctx.GetClock().Now().UTC().Add(time.Minute*10), subject, requester)
}
