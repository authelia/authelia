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
)

func handleOIDCAuthorizationConsentModeExplicit(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		consentID uuid.UUID
		err       error
	)

	bytesConsentID := ctx.QueryArgs().PeekBytes(queryArgConsentID)

	switch len(bytesConsentID) {
	case 0:
		return handleOIDCAuthorizationConsentGenerate(ctx, issuer, client, userSession, subject, rw, r, requester)
	default:
		if consentID, err = uuid.Parse(string(bytesConsentID)); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred parsing the challenge id: %+v", requester.GetID(), requester.GetClient().GetID(), err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModeExplicitWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModeExplicitWithID(ctx *middlewares.AutheliaCtx, _ *url.URL, _ *oidc.Client,
	userSession session.UserSession, subject uuid.UUID, consentID uuid.UUID,
	rw http.ResponseWriter, _ *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consentID.ID() == 0 {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred during consent session lookup: the consent id had a zero value", requester.GetID(), requester.GetClient().GetID())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consentID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred during consent session lookup: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if subject.ID() != consent.Subject.UUID.ID() {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred during consent session lookup: user '%s' with subject '%s' is not authorized to consent for subject '%s'", requester.GetID(), requester.GetClient().GetID(), userSession.Username, subject, consent.Subject.UUID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if !consent.CanGrant() {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred performing consent for consent session with id '%s': the session does not appear to be valid for implicit/explicit consent", requester.GetID(), requester.GetClient().GetID(), consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	if !consent.IsAuthorized() {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred performing consent for consent session with id '%s': the user did not provide their explicit consent", requester.GetID(), requester.GetClient().GetID(), consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	return consent, false
}
