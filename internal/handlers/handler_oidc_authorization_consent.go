package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func handleOIDCAuthorizationConsent(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		subject uuid.UUID
		err     error
	)

	var handler handlerAuthorizationConsent

	policy := client.GetAuthorizationPolicy()
	level := policy.GetRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()})

	switch {
	case userSession.IsAnonymous():
		handler = handleOIDCAuthorizationConsentNotAuthenticated
	case authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel, level):
		if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentCantGetSubject, requester.GetID(), client.GetID(), client.GetConsentPolicy(), userSession.Username, client.GetSectorIdentifierURI(), err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrSubjectCouldNotLookup)

			return nil, true
		}

		switch client.GetConsentPolicy().Mode {
		case oidc.ClientConsentModeExplicit:
			handler = handleOIDCAuthorizationConsentModeExplicit
		case oidc.ClientConsentModeImplicit:
			handler = handleOIDCAuthorizationConsentModeImplicit
		case oidc.ClientConsentModePreConfigured:
			handler = handleOIDCAuthorizationConsentModePreConfigured
		default:
			ctx.Logger.Errorf(logFmtErrConsentCantDetermineConsentMode, requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not determine the client consent mode."))

			return nil, true
		}
	case level == authorization.Denied:
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' using policy '%s' could not be processed: the user '%s' is not authorized to use this client", requester.GetID(), client.GetID(), policy.Name, userSession.Username)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

		return nil, true
	default:
		if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentCantGetSubject, requester.GetID(), client.GetID(), client.GetConsentPolicy(), userSession.Username, client.GetSectorIdentifierURI(), err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrSubjectCouldNotLookup)

			return nil, true
		}

		handler = handleOIDCAuthorizationConsentGenerate
	}

	return handler(ctx, issuer, client, userSession, subject, rw, r, requester)
}

func handleOIDCAuthorizationConsentNotAuthenticated(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	_ session.UserSession, _ uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var err error

	if consent, err = handleOpenIDConnectNewConsentSession(uuid.UUID{}, requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "saving", err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	redirectionURL := handleOIDCAuthorizationConsentGetRedirectionURL(ctx, issuer, consent)

	handleOIDCPushedAuthorizeConsent(ctx, requester, r.Form)

	http.Redirect(rw, r, redirectionURL.String(), http.StatusFound)

	return nil, true
}

func handleOIDCAuthorizationConsentGenerate(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	ctx.Logger.Debugf(logFmtDbgConsentGenerate, requester.GetID(), client.GetID(), client.GetConsentPolicy())

	if len(ctx.QueryArgs().PeekBytes(qryArgConsentID)) != 0 {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", errors.New("consent id value was present when it should be absent"))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if consent, err = handleOpenIDConnectNewConsentSession(subject, requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "saving", err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	if oidc.RequesterRequiresLogin(requester, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		handleOIDCAuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	} else {
		ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime(), "prompt": requester.GetRequestForm().Get("prompt")}).Debugf("Authorization Request with id '%s' on client with id '%s' is not being redirected for reauthentication", requester.GetID(), client.GetID())
	}

	handleOIDCAuthorizationConsentRedirect(ctx, issuer, consent, client, userSession, rw, r, requester)

	return consent, true
}

func handleOIDCAuthorizationConsentRedirect(ctx *middlewares.AutheliaCtx, issuer *url.URL, consent *model.OAuth2ConsentSession, client oidc.Client,
	userSession session.UserSession, rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) {
	var location *url.URL

	if client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		location, _ = url.ParseRequestURI(issuer.String())
		location.Path = path.Join(location.Path, oidc.EndpointPathConsentDecision)

		query := location.Query()
		query.Set(queryArgID, consent.ChallengeID.String())

		location.RawQuery = query.Encode()

		ctx.Logger.Debugf(logFmtDbgConsentAuthenticationSufficiency, requester.GetID(), client.GetID(), client.GetConsentPolicy(), userSession.AuthenticationLevel.String(), "sufficient", client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}))
	} else {
		location = handleOIDCAuthorizationConsentGetRedirectionURL(ctx, issuer, consent)

		ctx.Logger.Debugf(logFmtDbgConsentAuthenticationSufficiency, requester.GetID(), client.GetID(), client.GetConsentPolicy(), userSession.AuthenticationLevel.String(), "insufficient", client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}))
	}

	handleOIDCPushedAuthorizeConsent(ctx, requester, r.Form)

	ctx.Logger.Debugf(logFmtDbgConsentRedirect, requester.GetID(), client.GetID(), client.GetConsentPolicy(), location)

	http.Redirect(rw, r, location.String(), http.StatusFound)
}

func handleOIDCPushedAuthorizeConsent(ctx *middlewares.AutheliaCtx, requester oauthelia2.AuthorizeRequester, form url.Values) {
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

func handleOIDCAuthorizationConsentPromptLoginRedirect(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client, userSession session.UserSession, rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester, consent *model.OAuth2ConsentSession) {
	ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime()}).Debugf("Authorization Request with id '%s' on client with id '%s' is being redirected for reauthentication: prompt type login was requested", requester.GetID(), client.GetID())

	handleOIDCPushedAuthorizeConsent(ctx, requester, r.Form)

	redirectionURL := issuer.JoinPath(oidc.EndpointPathConsentLogin)

	query := redirectionURL.Query()
	query.Set(queryArgWorkflow, workflowOpenIDConnect)
	query.Set(queryArgWorkflowID, consent.ChallengeID.String())

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
	query.Set(queryArgWorkflow, workflowOpenIDConnect)
	query.Set(queryArgWorkflowID, consent.ChallengeID.String())

	redirectURL.RawQuery = query.Encode()

	return redirectURL
}

func handleOpenIDConnectNewConsentSession(subject uuid.UUID, requester oauthelia2.AuthorizeRequester, prefixPAR string) (consent *model.OAuth2ConsentSession, err error) {
	if oidc.IsPushedAuthorizedRequest(requester, prefixPAR) {
		form := url.Values{}

		form.Set(oidc.FormParameterRequestURI, requester.GetRequestForm().Get(oidc.FormParameterRequestURI))

		if requester.GetRequestForm().Has(oidc.FormParameterClientID) {
			form.Set(oidc.FormParameterClientID, requester.GetRequestForm().Get(oidc.FormParameterClientID))
		}

		return model.NewOAuth2ConsentSessionWithForm(subject, requester, form)
	}

	return model.NewOAuth2ConsentSession(subject, requester)
}

func verifyOIDCUserAuthorizedForConsent(ctx *middlewares.AutheliaCtx, client oidc.Client, userSession session.UserSession, consent *model.OAuth2ConsentSession, subject uuid.UUID) (err error) {
	if client == nil {
		if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
			return fmt.Errorf("failed to retrieve client: %w", err)
		}
	}

	if subject == uuid.Nil {
		if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			return fmt.Errorf("failed to lookup subject: %w", err)
		}
	}

	if !consent.Subject.Valid {
		if subject == uuid.Nil {
			return fmt.Errorf("the consent subject is null for consent session with id '%d' for anonymous user", consent.ID)
		}

		consent.Subject = model.NullUUID(subject)

		if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionSubject(ctx, consent); err != nil {
			return fmt.Errorf("failed to update the consent subject: %w", err)
		}
	}

	if consent.Subject.UUID != subject {
		return fmt.Errorf("the consent subject identifier '%s' isn't owned by user '%s' who has a subject identifier of '%s' with sector identifier '%s'", consent.Subject.UUID, userSession.Username, subject, client.GetSectorIdentifierURI())
	}

	return nil
}
