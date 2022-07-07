package handlers

import (
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
	"github.com/authelia/authelia/v4/internal/storage"
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

	if issuer, err = url.Parse(rootURI); err != nil {
		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not safely determine the issuer."))

		return nil, true
	}

	// This prevents the consent request from being generated until the authentication level is sufficient.
	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) || userSession.Username == "" {
		redirectURL := getOIDCAuthorizationRedirectURL(issuer, requester)

		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being redirected due to insufficient authentication", requester.GetID(), client.GetID())

		http.Redirect(rw, r, redirectURL.String(), http.StatusFound)

		return nil, true
	}

	if subject, err = ctx.Providers.OpenIDConnect.Store.GetSubject(ctx, client.GetSectorIdentifier(), userSession.Username); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retrieving subject identifier for user '%s' and sector identifier '%s': %+v", requester.GetID(), client.GetID(), userSession.Username, client.GetSectorIdentifier(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not retrieve the subject."))

		return nil, true
	}

	var consentIDBytes []byte

	if consentIDBytes = ctx.QueryArgs().Peek("consent_id"); len(consentIDBytes) != 0 {
		var consentID uuid.UUID

		if consentID, err = uuid.Parse(string(consentIDBytes)); err != nil {
			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Consent Session ID was Malformed."))

			return nil, true
		}

		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' proceeding to lookup consent by challenge id '%s'", requester.GetID(), client.GetID(), consentID)

		return handleOIDCAuthorizationConsentWithChallengeID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}

	return handleOIDCAuthorizationConsentGenerate(ctx, issuer, client, userSession, subject, rw, r, requester)
}

func handleOIDCAuthorizationConsentWithChallengeID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject, challengeID uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, challengeID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred during consent session lookup: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Failed to lookup consent session."))

		return nil, true
	}

	if consent.Subject.UUID != subject || !consent.Subject.Valid || consent.Subject.UUID.ID() == 0 {
		ctx.Logger.Debugf("Consent Subject: %s (%d) Valid %t, Subject: %s (%d)", consent.Subject, consent.Subject.ID(), consent.Subject.Valid, subject, subject.ID())
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not process consent session with challenge id '%s': the user '%s' is not authorized to process consent sessions for subject '%s'", requester.GetID(), client.GetID(), consent.ChallengeID, userSession.Username, consent.Subject.UUID)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("The user is not authorized to perform consent."))

		return nil, true
	}

	if consent.Responded() {
		if consent.Granted {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: this consent session with challenge id '%s' was already granted", requester.GetID(), client.GetID(), consent.ChallengeID)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Authorization already granted."))

			return nil, true
		}

		ctx.Logger.Debugf("Authorization Request with id '%s' loaded consent session with id '%d' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s'", requester.GetID(), consent.ID, consent.ChallengeID, client.GetID(), consent.Subject, strings.Join(requester.GetRequestedScopes(), " "))

		if consent.IsDenied() {
			ctx.Logger.Warnf("Authorization Request with id '%s' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s' was not denied by the user durng the consent session", requester.GetID(), consent.ChallengeID, client.GetID(), consent.Subject, strings.Join(requester.GetRequestedScopes(), " "))

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrAccessDenied)

			return nil, true
		}

		return consent, false
	}

	handleOIDCAuthorizationConsentRedirect(ctx, issuer, consent, client, userSession, rw, r, requester)

	return consent, true
}

func handleOIDCAuthorizationConsentGenerate(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	scopes, audience := getOIDCExpectedScopesAndAudienceFromRequest(requester)

	if consent, err = getOIDCPreConfiguredConsent(ctx, client.GetID(), subject, scopes, audience); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' had error looking up pre-configured consent sessions: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not lookup the consent session."))

		return nil, true
	}

	if consent != nil {
		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' successfully looked up pre-configured consent with challenge id '%s'", requester.GetID(), client.GetID(), consent.ChallengeID)

		return consent, false
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' proceeding to generate a new consent due to unsuccessful lookup of pre-configured consent", requester.GetID(), client.GetID())

	if consent, err = model.NewOAuth2ConsentSession(model.NullUUID{UUID: subject, Valid: true}, requester); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not generate the consent session."))

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, *consent); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the consent session."))

		return nil, true
	}

	handleOIDCAuthorizationConsentRedirect(ctx, issuer, consent, client, userSession, rw, r, requester)

	return consent, true
}

func handleOIDCAuthorizationConsentRedirect(ctx *middlewares.AutheliaCtx, issuer *url.URL, consent *model.OAuth2ConsentSession, client *oidc.Client,
	userSession session.UserSession, rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) {
	var location *url.URL

	if client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		location, _ = url.Parse(issuer.String())
		location.Path = path.Join(location.Path, "/consent")

		query := location.Query()
		query.Set("consent_id", consent.ChallengeID.String())

		location.RawQuery = query.Encode()

		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' authentication level '%s' is sufficient for client level '%s'", requester.GetID(), client.GetID(), authentication.LevelToString(userSession.AuthenticationLevel), authorization.LevelToString(client.Policy))
	} else {
		location = getOIDCAuthorizationRedirectURL(issuer, requester)

		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' authentication level '%s' is insufficient for client level '%s'", requester.GetID(), client.GetID(), authentication.LevelToString(userSession.AuthenticationLevel), authorization.LevelToString(client.Policy))
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being redirected to '%s'", requester.GetID(), client.GetID(), location)

	http.Redirect(rw, r, location.String(), http.StatusFound)
}

func getOIDCAuthorizationRedirectURL(issuer *url.URL, requester fosite.AuthorizeRequester) (redirectURL *url.URL) {
	redirectURL, _ = url.Parse(issuer.String())

	authorizationURL, _ := url.Parse(issuer.String())

	authorizationURL.Path = path.Join(authorizationURL.Path, oidc.AuthorizationPath)
	authorizationURL.RawQuery = requester.GetRequestForm().Encode()

	query := redirectURL.Query()
	query.Set("rd", authorizationURL.String())
	query.Set("workflow", "openid_connect")

	redirectURL.RawQuery = query.Encode()

	return redirectURL
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

func getOIDCPreConfiguredConsent(ctx *middlewares.AutheliaCtx, clientID string, subject uuid.UUID, scopes, audience []string) (consent *model.OAuth2ConsentSession, err error) {
	var (
		rows *storage.ConsentSessionRows
	)

	ctx.Logger.Debugf("Consent Session is being checked for pre-configuration with signature of client id '%s' and subject '%s'", clientID, subject)

	if rows, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionsPreConfigured(ctx, clientID, subject); err != nil {
		ctx.Logger.Debugf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' failed with error during load: %+v", clientID, subject, err)

		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			ctx.Logger.Errorf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' failed to close rows with error: %+v", clientID, subject, err)
		}
	}()

	for rows.Next() {
		if consent, err = rows.Get(); err != nil {
			ctx.Logger.Debugf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' failed with error during iteration: %+v", clientID, subject, err)

			return nil, err
		}

		if consent.HasExactGrants(scopes, audience) && consent.CanGrant() {
			break
		}
	}

	if consent != nil && consent.HasExactGrants(scopes, audience) && consent.CanGrant() {
		ctx.Logger.Debugf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' found a result with challenge id '%s'", clientID, subject, consent.ChallengeID)

		return consent, nil
	}

	ctx.Logger.Debugf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' did not find any results", clientID, subject)

	return nil, nil
}
