package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

func handleOIDCAuthorizationConsent(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	userSession session.UserSession,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		issuer  *url.URL
		subject uuid.UUID
		err     error
	)

	if issuer, err = url.ParseRequestURI(rootURI); err != nil {
		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrIssuerCouldNotDerive)

		return nil, true
	}

	if !strings.HasSuffix(issuer.Path, "/") {
		issuer.Path += "/"
	}

	var handler handlerAuthorizationConsent

	switch {
	case userSession.AuthenticationLevel == authentication.NotAuthenticated || userSession.Username == "":
		handler = handleOIDCAuthorizationConsentNotAuthenticated
	case client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel):
		if subject, err = ctx.Providers.OpenIDConnect.Store.GetSubject(ctx, client.GetSectorIdentifier(), userSession.Username); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retrieving subject identifier for user '%s' and sector identifier '%s': %+v", requester.GetID(), client.GetID(), userSession.Username, client.GetSectorIdentifier(), err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrSubjectCouldNotLookup)

			return nil, true
		}

		switch client.Consent.Mode {
		case oidc.ClientConsentModeExplicit:
			handler = handleOIDCAuthorizationConsentModeExplicit
		case oidc.ClientConsentModeImplicit:
			handler = handleOIDCAuthorizationConsentModeImplicit
		case oidc.ClientConsentModePreConfigured:
			handler = handleOIDCAuthorizationConsentModePreConfigured
		default:
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: client consent mode could not be reliably determined", requester.GetID(), requester.GetClient().GetID())

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not determine the client consent mode."))

			return nil, true
		}
	default:
		handler = handleOIDCAuthorizationConsentGenerate
	}

	return handler(ctx, issuer, client, userSession, subject, rw, r, requester)
}

func handleOIDCAuthorizationConsentNotAuthenticated(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	redirectionURL := handleOIDCAuthorizationConsentGetRedirectionURL(issuer, nil, requester)

	http.Redirect(rw, r, redirectionURL.String(), http.StatusFound)

	return nil, true
}

func handleOIDCAuthorizationConsentGenerate(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' proceeding to generate a new consent session", requester.GetID(), client.GetID())

	if len(ctx.QueryArgs().PeekBytes(queryArgConsentID)) != 0 {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: consent id value was present when it should be absent", requester.GetID(), requester.GetClient().GetID())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if consent, err = model.NewOAuth2ConsentSession(subject, requester); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, *consent); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	handleOIDCAuthorizationConsentRedirect(ctx, issuer, consent, client, userSession, rw, r, requester)

	return consent, true
}

func handleOIDCAuthorizationConsentRedirect(ctx *middlewares.AutheliaCtx, issuer *url.URL, consent *model.OAuth2ConsentSession, client *oidc.Client,
	userSession session.UserSession, rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) {
	var location *url.URL

	if client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		location, _ = url.ParseRequestURI(issuer.String())
		location.Path = path.Join(location.Path, oidc.EndpointConsent)

		query := location.Query()
		query.Set(queryArgStrID, consent.ChallengeID.String())

		location.RawQuery = query.Encode()

		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' authentication level '%s' is sufficient for client level '%s'", requester.GetID(), client.GetID(), authentication.LevelToString(userSession.AuthenticationLevel), authorization.LevelToString(client.Policy))
	} else {
		location = handleOIDCAuthorizationConsentGetRedirectionURL(issuer, consent, requester)

		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' authentication level '%s' is insufficient for client level '%s'", requester.GetID(), client.GetID(), authentication.LevelToString(userSession.AuthenticationLevel), authorization.LevelToString(client.Policy))
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being redirected to '%s'", requester.GetID(), client.GetID(), location)

	http.Redirect(rw, r, location.String(), http.StatusFound)
}

func handleOIDCAuthorizationConsentGetRedirectionURL(issuer *url.URL, consent *model.OAuth2ConsentSession, requester fosite.AuthorizeRequester) (redirectURL *url.URL) {
	redirectURL, _ = url.ParseRequestURI(issuer.String())

	query := redirectURL.Query()
	query.Set(queryArgStrWorkflow, workflowOpenIDConnect)

	switch {
	case consent != nil:
		query.Set(queryArgStrWorkflowID, consent.ChallengeID.String())
	case requester != nil:
		rd, _ := url.ParseRequestURI(issuer.String())
		rd.Path = path.Join(rd.Path, oidc.EndpointAuthorization)
		rd.RawQuery = requester.GetRequestForm().Encode()

		query.Set("rd", rd.String())
	}

	redirectURL.RawQuery = query.Encode()

	return redirectURL
}

func verifyOIDCUserAuthorizedForConsent(ctx *middlewares.AutheliaCtx, client *oidc.Client, userSession session.UserSession, consent *model.OAuth2ConsentSession, subject uuid.UUID) (err error) {
	var sid uint32

	if client == nil {
		if client, err = ctx.Providers.OpenIDConnect.Store.GetFullClient(consent.ClientID); err != nil {
			return fmt.Errorf("failed to retrieve client: %w", err)
		}
	}

	if sid = subject.ID(); sid == 0 {
		if subject, err = ctx.Providers.OpenIDConnect.Store.GetSubject(ctx, client.GetSectorIdentifier(), userSession.Username); err != nil {
			return fmt.Errorf("failed to lookup subject: %w", err)
		}

		sid = subject.ID()
	}

	if !consent.Subject.Valid {
		if sid == 0 {
			return fmt.Errorf("the consent subject is null for consent session with id '%d' for anonymous user", consent.ID)
		}

		consent.Subject = uuid.NullUUID{UUID: subject, Valid: true}

		if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionSubject(ctx, *consent); err != nil {
			return fmt.Errorf("failed to update the consent subject: %w", err)
		}
	}

	if consent.Subject.UUID.ID() != sid {
		return fmt.Errorf("the consent subject identifier '%s' isn't owned by user '%s' who has a subject identifier of '%s' with sector identifier '%s'", consent.Subject.UUID, userSession.Username, subject, client.GetSectorIdentifier())
	}

	return nil
}

func getOIDCExpectedScopesAndAudienceFromRequest(requester fosite.Requester) (scopes, audience []string) {
	return getOIDCExpectedScopesAndAudience(requester.GetClient().GetID(), requester.GetRequestedScopes(), requester.GetRequestedAudience())
}

func getOIDCExpectedScopesAndAudience(clientID string, scopes, audience []string) (expectedScopes, expectedAudience []string) {
	if !utils.IsStringInSlice(clientID, audience) {
		audience = append(audience, clientID)
	}

	return scopes, audience
}

const (
	logFmtAuthorizationPrefix = "Authorization Request with id '%s' on client with id '%s' "
	logFmtConsentPrefix       = logFmtAuthorizationPrefix + "using consent mode '%s' "

	logFmtErrConsentParseChallengeID = logFmtConsentPrefix + "could not be processed: error occurred parsing the consent id (challenge) '%s': %+v"
	logFmtErrConsentPreConfLookup    = logFmtConsentPrefix + "had error looking up pre-configured consent sessions: %+v"
	logFmtErrConsentPreConfRowsClose = logFmtConsentPrefix + "had error closing rows while looking up pre-configured consent sessions: %+v"

	logFmtErrConsentWithIDCouldNotBeProcessed   = logFmtConsentPrefix + "could not be processed: error occurred performing consent for consent session with id '%s': "
	logFmtErrConsentNotValidForPreConf          = logFmtErrConsentWithIDCouldNotBeProcessed + "the session does not appear to be valid for pre-configured consent"
	logFmtErrConsentNotProvided                 = logFmtErrConsentWithIDCouldNotBeProcessed + "the user did not provide their consent"
	logFmtErrConsentZeroID                      = logFmtConsentPrefix + "could not be processed: the consent id had a zero value"
	logFmtErrConsentLookupLoadingSession        = logFmtErrConsentWithIDCouldNotBeProcessed + "error occurred while loading session: %+v"
	logFmtErrConsentSessionSubjectNotAuthorized = logFmtErrConsentWithIDCouldNotBeProcessed + "user '%s' with subject '%s' is not authorized to consent for subject '%s'"
	logFmtErrConsentCantGrant                   = logFmtErrConsentWithIDCouldNotBeProcessed + "the session does not appear to be valid for %s consent: either the subject is null, the consent has already been granted, or the consent session is a pre-configured session"
	logFmtErrConsentCantGrantPreConf            = logFmtErrConsentWithIDCouldNotBeProcessed + "the session does not appear to be valid for pre-configured consent: either the subject is null, the consent has been granted and is either not pre-configured, or the pre-configuration is expired"
	logFmtErrConsentCantGrantNotResponded       = logFmtErrConsentWithIDCouldNotBeProcessed + "the user did not provide their explicit consent"
	logFmtErrConsentCantGrantRejected           = logFmtErrConsentWithIDCouldNotBeProcessed + "the user explicitly rejected this consent session"
	logFmtErrConsentSaveSessionResponse         = logFmtErrConsentWithIDCouldNotBeProcessed + "error occurred saving consent session response: %+v"
	logFmtErrConsentSaveSession                 = logFmtErrConsentWithIDCouldNotBeProcessed + "error occurred saving consent session: %+v"
	logFmtErrConsentGenerate                    = logFmtConsentPrefix + "could not be processed: error occurred generating consent: %+v"

	logFmtDbgConsentPreConfSuccessfulLookup   = logFmtConsentPrefix + "successfully looked up pre-configured consent with signature of client id '%s' and subject '%s' and scopes '%s' with challenge id '%s'"
	logFmtDbgConsentPreConfSuccessfulAndValid = logFmtConsentPrefix + "successfully looked up and validated pre-configured consent with challenge id '%s'"
	logFmtDbgConsentPreConfUnsuccessfulLookup = logFmtConsentPrefix + "unsuccessfully looked up pre-configured consent with signature of client id '%s' and subject '%s' and scopes '%s'"
	logFmtDbgConsentPreConfTryingLookup       = logFmtConsentPrefix + "attempting to discover pre-configurations with signature of client id '%s' and subject '%s' and scopes '%s'"
)
