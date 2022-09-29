package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
			ctx.Logger.Errorf(logFmtErrConsentParseChallengeID, requester.GetID(), client.GetID(), client.Consent, bytesConsentID, err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModePreConfiguredWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModePreConfiguredWithID(ctx *middlewares.AutheliaCtx, _ *url.URL, client *oidc.Client,
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

	if !consent.CanGrantConsentModePreConfigured() {
		ctx.Logger.Errorf(logFmtErrConsentCantGrantPreConf, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID)

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

func handleOIDCAuthorizationConsentModePreConfiguredWithoutID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consent, err = handleOIDCAuthorizationConsentModePreConfiguredGetConsent(ctx, client, subject, requester); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentPreConfLookup, requester.GetID(), client.GetID(), client.Consent, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if consent == nil {
		return handleOIDCAuthorizationConsentGenerate(ctx, issuer, client, userSession, subject, rw, r, requester)
	}

	if !consent.CanGrantConsentModePreConfigured() {
		ctx.Logger.Errorf(logFmtErrConsentNotValidForPreConf, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	if !consent.IsAuthorized() {
		ctx.Logger.Errorf(logFmtErrConsentNotProvided, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	ctx.Logger.Debugf(logFmtDbgConsentPreConfSuccessfulAndValid, requester.GetID(), client.GetID(), client.Consent, consent.ChallengeID)

	return consent, false
}

func handleOIDCAuthorizationConsentModePreConfiguredGetConsent(ctx *middlewares.AutheliaCtx, client *oidc.Client, subject uuid.UUID, requester fosite.Requester) (consent *model.OAuth2ConsentSession, err error) {
	var (
		rows *storage.ConsentSessionRows
	)

	ctx.Logger.Debugf(logFmtDbgConsentPreConfTryingLookup, requester.GetID(), client.GetID(), client.Consent, client.GetID(), subject, strings.Join(requester.GetRequestedScopes(), " "))

	if rows, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionsPreConfigured(ctx, client.GetID(), subject); err != nil {
		return nil, fmt.Errorf("error loading rows: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentPreConfRowsClose, requester.GetID(), client.GetID(), client.Consent, err)
		}
	}()

	scopes, audience := getOIDCExpectedScopesAndAudienceFromRequest(requester)

	for rows.Next() {
		if consent, err = rows.Get(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %w", err)
		}

		if consent.HasExactGrants(scopes, audience) && consent.CanGrantZ() {
			break
		}
	}

	if consent != nil && consent.HasExactGrants(scopes, audience) && consent.CanGrantZ() {
		ctx.Logger.Debugf(logFmtDbgConsentPreConfSuccessfulLookup, requester.GetID(), client.GetID(), client.Consent, client.GetID(), subject, strings.Join(requester.GetRequestedScopes(), " "), consent.ChallengeID)

		return consent, nil
	}

	ctx.Logger.Debugf(logFmtDbgConsentPreConfUnsuccessfulLookup, requester.GetID(), client.GetID(), client.Consent, client.GetID(), subject, strings.Join(requester.GetRequestedScopes(), " "))

	return nil, nil
}
