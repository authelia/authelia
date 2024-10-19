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
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
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

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModeImplicitWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModeImplicitWithID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID, consentID uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consentID == uuid.Nil {
		ctx.Logger.Errorf(logFmtErrConsentZeroID, requester.GetID(), client.GetID(), client.GetConsentPolicy())

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consentID); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentLookupLoadingSession, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consentID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if subject != consent.Subject.UUID {
		ctx.Logger.Errorf(logFmtErrConsentSessionSubjectNotAuthorized, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, userSession.Username, subject, consent.Subject.UUID)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if oidc.RequesterRequiresLogin(requester, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		handleOIDCAuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	} else {
		ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime(), "prompt": requester.GetRequestForm().Get("prompt")}).Debugf("Authorization Request with id '%s' on client with id '%s' is not being redirected for reauthentication", requester.GetID(), client.GetID())
	}

	if !consent.CanGrant() {
		ctx.Logger.Errorf(logFmtErrConsentCantGrant, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, "implicit")

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	consent.Grant()

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}

func handleOIDCAuthorizationConsentModeImplicitWithoutID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consent, err = handleOpenIDConnectNewConsentSession(subject, requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentGenerateError, requester.GetID(), client.GetID(), client.GetConsentPolicy(), "generating", err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotGenerate)

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSession, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	challenge := consent.ChallengeID

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, challenge); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSession, requester.GetID(), client.GetID(), client.GetConsentPolicy(), challenge, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	if oidc.RequesterRequiresLogin(requester, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		handleOIDCAuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	} else {
		ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime(), "prompt": requester.GetRequestForm().Get("prompt")}).Debugf("Authorization Request with id '%s' on client with id '%s' is not being redirected for reauthentication", requester.GetID(), client.GetID())
	}

	consent.Grant()

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}
