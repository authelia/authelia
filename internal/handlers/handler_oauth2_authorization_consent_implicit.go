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

func handleOAuth2AuthorizationConsentModeImplicit(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		consentID uuid.UUID
		err       error
	)

	switch bytesConsentID := ctx.QueryArgs().PeekBytes(qryArgConsentID); len(bytesConsentID) {
	case 0:
		return handleOAuth2AuthorizationConsentModeImplicitWithoutID(ctx, issuer, client, userSession, subject, rw, r, requester)
	default:
		if consentID, err = uuid.ParseBytes(bytesConsentID); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentParseChallengeID, requester.GetID(), client.GetID(), client.GetConsentPolicy(), bytesConsentID, err)

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOAuth2AuthorizationConsentModeImplicitWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOAuth2AuthorizationConsentModeImplicitWithID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
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

	if (consent.Subject.UUID != uuid.Nil && subject != consent.Subject.UUID) || subject == uuid.Nil {
		ctx.Logger.Errorf(logFmtErrConsentSessionSubjectNotAuthorized, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, userSession.Username, subject, consent.Subject.UUID)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if oidc.RequesterRequiresLogin(requester, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		handleOAuth2AuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	} else {
		ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime(), "prompt": requester.GetRequestForm().Get("prompt")}).Debugf("Authorization Request with id '%s' on client with id '%s' is not being redirected for reauthentication", requester.GetID(), client.GetID())
	}

	if !consent.CanGrant(ctx.GetClock().Now()) {
		ctx.Logger.Errorf(logFmtErrConsentCantGrant, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, "implicit")

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	var requests *oidc.ClaimsRequests

	if requests, err = oidc.NewClaimRequests(requester.GetRequestForm()); err == nil && requests != nil {
		oidc.ConsentGrantImplicit(consent, requests.ToSlice(), subject, ctx.GetClock().Now())
	} else {
		oidc.ConsentGrantImplicit(consent, nil, subject, ctx.GetClock().Now())
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}

func handleOAuth2AuthorizationConsentModeImplicitWithoutID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)
	if consent, err = handleOAuth2NewConsentSession(ctx, subject, requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
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
		handleOAuth2AuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	} else {
		ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime(), "prompt": requester.GetRequestForm().Get("prompt")}).Debugf("Authorization Request with id '%s' on client with id '%s' is not being redirected for reauthentication", requester.GetID(), client.GetID())
	}

	var requests *oidc.ClaimsRequests

	if requests, err = oidc.NewClaimRequests(requester.GetRequestForm()); err == nil && requests != nil {
		oidc.ConsentGrantImplicit(consent, requests.ToSlice(), subject, ctx.GetClock().Now())
	} else {
		oidc.ConsentGrantImplicit(consent, nil, subject, ctx.GetClock().Now())
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}
