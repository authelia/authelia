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
			ctx.Logger.Errorf(logFmtErrConsentParseChallengeID, requester.GetID(), client.GetID(), client.Consent, bytesConsentID, err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModeExplicitWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModeExplicitWithID(ctx *middlewares.AutheliaCtx, _ *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID, consentID uuid.UUID,
	rw http.ResponseWriter, _ *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consentID.ID() == 0 {
		ctx.Logger.Errorf(logFmtErrConsentZeroID, requester.GetID(), client.GetID(), client.Consent)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consentID); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentLookupLoadingSession, requester.GetID(), client.GetID(), client.Consent, consentID, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if subject.ID() != consent.Subject.UUID.ID() {
		ctx.Logger.Errorf(logFmtErrConsentSessionSubjectNotAuthorized, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID, userSession.Username, subject, consent.Subject.UUID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if !consent.CanGrant() {
		ctx.Logger.Errorf(logFmtErrConsentCantGrant, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID, "explicit")

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	if !consent.IsAuthorized() {
		if consent.Responded() {
			ctx.Logger.Errorf(logFmtErrConsentCantGrantRejected, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

			return nil, true
		}

		ctx.Logger.Errorf(logFmtErrConsentCantGrantNotResponded, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	return consent, false
}
