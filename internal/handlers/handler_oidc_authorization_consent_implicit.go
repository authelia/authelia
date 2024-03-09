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

func handleOIDCAuthorizationConsentModeImplicit(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		consentID uuid.UUID
		err       error
	)

	switch bytesConsentID := ctx.QueryArgs().PeekBytes(qryArgConsentID); len(bytesConsentID) {
	case 0:
		return handleOIDCAuthorizationConsentModeImplicitWithoutID(ctx, issuer, client, userSession, subject, rw, r, requester)
	default:
		if consentID, err = uuid.ParseBytes(bytesConsentID); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentParseChallengeID, requester.GetID(), client.GetID(), client.GetConsentPolicy(), bytesConsentID, err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModeImplicitWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModeImplicitWithID(ctx *middlewares.AutheliaCtx, _ *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID, consentID uuid.UUID,
	rw http.ResponseWriter, _ *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
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
		ctx.Logger.Errorf(logFmtErrConsentCantGrant, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, "implicit")

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	consent.Grant()

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, *consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}

func handleOIDCAuthorizationConsentModeImplicitWithoutID(ctx *middlewares.AutheliaCtx, _ *url.URL, client oidc.Client,
	_ session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, _ *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consent, err = handleOpenIDConnectNewConsentSession(subject, requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, *consent); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSession, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consent.ChallengeID); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSession, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	consent.Grant()

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, *consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}
