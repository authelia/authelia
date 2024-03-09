package handlers

import (
	"net/http"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func handleOIDCAuthorizationConsentModeExplicit(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		consentID uuid.UUID
		err       error
	)

	bytesConsentID := ctx.QueryArgs().PeekBytes(qryArgConsentID)

	switch len(bytesConsentID) {
	case 0:
		return handleOIDCAuthorizationConsentGenerate(ctx, issuer, client, userSession, subject, rw, r, requester)
	default:
		if consentID, err = uuid.ParseBytes(bytesConsentID); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentParseChallengeID, requester.GetID(), client.GetID(), client.GetConsentPolicy(), bytesConsentID, err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModeExplicitWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModeExplicitWithID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID, consentID uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consentID.ID() == 0 {
		ctx.Logger.Errorf(logFmtErrConsentZeroID, requester.GetID(), client.GetID(), client.GetConsentPolicy())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consentID); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentLookupLoadingSession, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consentID, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if subject.ID() != consent.Subject.UUID.ID() {
		ctx.Logger.Errorf(logFmtErrConsentSessionSubjectNotAuthorized, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, userSession.Username, subject, consent.Subject.UUID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if !consent.CanGrant() {
		ctx.Logger.Errorf(logFmtErrConsentCantGrant, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, "explicit")

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	if !consent.IsAuthorized() {
		if consent.Responded() {
			ctx.Logger.Errorf(logFmtErrConsentCantGrantRejected, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied)

			return nil, true
		}

		handleOIDCAuthorizationConsentRedirect(ctx, issuer, consent, client, userSession, rw, r, requester)

		return nil, true
	}

	return consent, false
}
