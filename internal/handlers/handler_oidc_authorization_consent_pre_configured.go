package handlers

import (
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func handleOIDCAuthorizationConsentModePreConfigured(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		consentID uuid.UUID
		err       error
	)

	bytesConsentID := ctx.QueryArgs().PeekBytes(queryArgConsentID)

	switch len(bytesConsentID) {
	case 0:
		return handleOIDCAuthorizationConsentModePreConfiguredWithoutID(ctx, issuer, client, userSession, subject, rw, r, requester)
	default:
		if consentID, err = uuid.Parse(string(bytesConsentID)); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred parsing the challenge id: %+v", requester.GetID(), requester.GetClient().GetID(), err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModeExplicitWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModePreConfiguredWithoutID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consent, err = handleOIDCAuthorizationConsentModePreConfiguredGetConsent(ctx, client.GetID(), subject, requester); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' had error looking up pre-configured consent sessions: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if consent == nil {
		return handleOIDCAuthorizationConsentGenerate(ctx, issuer, client, userSession, subject, rw, r, requester)
	}

	if !consent.CanGrantConsentModePreConfigured() {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred performing consent for consent session with id '%s': the session does not appear to be valid for pre-configured consent", requester.GetID(), requester.GetClient().GetID(), consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	if !consent.IsAuthorized() {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred performing consent for consent session with id '%s': the user did not provide their explicit consent", requester.GetID(), requester.GetClient().GetID(), consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' successfully looked up pre-configured consent with challenge id '%s'", requester.GetID(), client.GetID(), consent.ChallengeID)

	return consent, false
}

func handleOIDCAuthorizationConsentModePreConfiguredGetConsent(ctx *middlewares.AutheliaCtx, clientID string, subject uuid.UUID, requester fosite.Requester) (consent *model.OAuth2ConsentSession, err error) {
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

	scopes, audience := getOIDCExpectedScopesAndAudienceFromRequest(requester)

	for rows.Next() {
		if consent, err = rows.Get(); err != nil {
			ctx.Logger.Debugf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' failed with error during iteration: %+v", clientID, subject, err)

			return nil, err
		}

		if consent.HasExactGrants(scopes, audience) && consent.CanGrantZ() {
			break
		}
	}

	if consent != nil && consent.HasExactGrants(scopes, audience) && consent.CanGrantZ() {
		ctx.Logger.Debugf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' found a result with challenge id '%s'", clientID, subject, consent.ChallengeID)

		return consent, nil
	}

	ctx.Logger.Debugf("Consent Session checked for pre-configuration with signature of client id '%s' and subject '%s' did not find any results", clientID, subject)

	return nil, nil
}
