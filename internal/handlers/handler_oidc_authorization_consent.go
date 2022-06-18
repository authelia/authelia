package handlers

import (
	"fmt"
	"net/http"
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
	userSession session.UserSession, subject model.NullUUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	if userSession.ConsentChallengeID != nil {
		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' proceeding to lookup consent by challenge id '%s'", requester.GetID(), client.GetID(), userSession.ConsentChallengeID)

		return handleOIDCAuthorizationConsentWithChallengeID(ctx, rootURI, client, userSession, rw, r, requester)
	}

	if !subject.Valid {
		return handleOIDCAuthorizationConsentGenerate(ctx, rootURI, client, userSession, subject, rw, r, requester)
	}

	return handleOIDCAuthorizationConsentOrGenerate(ctx, rootURI, client, userSession, subject, rw, r, requester)
}

func handleOIDCAuthorizationConsentWithChallengeID(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	userSession session.UserSession,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, *userSession.ConsentChallengeID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred during consent session lookup: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Failed to lookup consent session."))

		userSession.ConsentChallengeID = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred unlinking consent session challenge id: %+v", requester.GetID(), requester.GetClient().GetID(), err)
		}

		return nil, true
	}

	if !consent.Subject.Valid {
		if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.Store.GetSubject(ctx, client.GetSectorIdentifier(), userSession.Username); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retrieving subject for user '%s': %+v", requester.GetID(), client.GetID(), userSession.Username, err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not retrieve the subject."))

			return nil, true
		}

		consent.Subject.Valid = true

		if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionSubject(ctx, *consent); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred updating consent session subject for user '%s': %+v", requester.GetID(), client.GetID(), userSession.Username, err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not update the consent session subject."))

			return nil, true
		}
	}

	if consent.Responded() {
		userSession.ConsentChallengeID = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving session: %+v", requester.GetID(), client.GetID(), err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the session."))

			return nil, true
		}

		if consent.Granted {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: this consent session with challenge id '%s' was already granted", requester.GetID(), client.GetID(), consent.ChallengeID.String())

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Authorization already granted."))

			return nil, true
		}

		ctx.Logger.Debugf("Authorization Request with id '%s' loaded consent session with id '%d' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s'", requester.GetID(), consent.ID, consent.ChallengeID.String(), client.GetID(), consent.Subject.String(), strings.Join(requester.GetRequestedScopes(), " "))

		if consent.IsDenied() {
			ctx.Logger.Warnf("Authorization Request with id '%s' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s' was not denied by the user durng the consent session", requester.GetID(), consent.ChallengeID.String(), client.GetID(), consent.Subject.String(), strings.Join(requester.GetRequestedScopes(), " "))

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrAccessDenied)

			return nil, true
		}

		return consent, false
	}

	handleOIDCAuthorizationConsentRedirect(ctx, rootURI, client, userSession, rw, r, requester)

	return consent, true
}

func handleOIDCAuthorizationConsentOrGenerate(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	userSession session.UserSession, subject model.NullUUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	scopes, audience := getOIDCExpectedScopesAndAudienceFromRequest(requester)

	if consent, err = getOIDCPreConfiguredConsent(ctx, client.GetID(), subject.UUID, scopes, audience); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' had error looking up pre-configured consent sessions: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not lookup the consent session."))

		return nil, true
	}

	if consent != nil {
		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' successfully looked up pre-configured consent with challenge id '%s'", requester.GetID(), client.GetID(), consent.ChallengeID)

		return consent, false
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' proceeding to generate a new consent due to unsuccessful lookup of pre-configured consent", requester.GetID(), client.GetID())

	return handleOIDCAuthorizationConsentGenerate(ctx, rootURI, client, userSession, subject, rw, r, requester)
}

func handleOIDCAuthorizationConsentGenerate(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	userSession session.UserSession, subject model.NullUUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var err error

	if consent, err = model.NewOAuth2ConsentSession(subject, requester); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not generate the consent session."))

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, *consent); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the consent session."))

		return nil, true
	}

	userSession.ConsentChallengeID = &consent.ChallengeID

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving user session for consent: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the user session."))

		return nil, true
	}

	handleOIDCAuthorizationConsentRedirect(ctx, rootURI, client, userSession, rw, r, requester)

	return consent, true
}

func handleOIDCAuthorizationConsentRedirect(ctx *middlewares.AutheliaCtx, destination string, client *oidc.Client,
	userSession session.UserSession, rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) {
	if client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' authentication level '%s' is sufficient for client level '%s'", requester.GetID(), client.GetID(), authentication.LevelToString(userSession.AuthenticationLevel), authorization.LevelToPolicy(client.Policy))

		destination = fmt.Sprintf("%s/consent", destination)
	} else {
		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' authentication level '%s' is insufficient for client level '%s'", requester.GetID(), client.GetID(), authentication.LevelToString(userSession.AuthenticationLevel), authorization.LevelToPolicy(client.Policy))
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being redirected to '%s'", requester.GetID(), client.GetID(), destination)

	http.Redirect(rw, r, destination, http.StatusFound)
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

func getOIDCPreConfiguredConsentFromClientAndConsent(ctx *middlewares.AutheliaCtx, client fosite.Client, consent *model.OAuth2ConsentSession) (preConfigConsent *model.OAuth2ConsentSession, err error) {
	if consent == nil || !consent.Subject.Valid {
		return nil, fmt.Errorf("invalid consent provided for pre-configured consent lookup")
	}

	scopes, audience := getOIDCExpectedScopesAndAudience(client.GetID(), consent.RequestedScopes, consent.RequestedAudience)

	// We can skip this error as it's handled at the authorization endpoint.
	preConfigConsent, _ = getOIDCPreConfiguredConsent(ctx, client.GetID(), consent.Subject.UUID, scopes, audience)

	return preConfigConsent, nil
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
