package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func handleOIDCAuthorizationConsentModePreConfigured(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		consentID uuid.UUID
		err       error
	)

	bytesConsentID := ctx.QueryArgs().PeekBytes(qryArgConsentID)

	switch len(bytesConsentID) {
	case 0:
		return handleOIDCAuthorizationConsentModePreConfiguredWithoutID(ctx, issuer, client, userSession, subject, rw, r, requester)
	default:
		if consentID, err = uuid.ParseBytes(bytesConsentID); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentParseChallengeID, requester.GetID(), client.GetID(), client.GetConsentPolicy(), bytesConsentID, err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOIDCAuthorizationConsentModePreConfiguredWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOIDCAuthorizationConsentModePreConfiguredWithID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID, consentID uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		config *model.OAuth2ConsentPreConfig
		err    error
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
		ctx.Logger.Errorf(logFmtErrConsentCantGrantPreConf, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	if config, err = handleOIDCAuthorizationConsentModePreConfiguredGetPreConfig(ctx, client, subject, requester); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentPreConfLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if config != nil {
		consent.Grant()

		consent.PreConfiguration = sql.NullInt64{Int64: config.ID, Valid: true}

		if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, *consent, false); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

			return nil, true
		}

		return consent, false
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

	if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptNone {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: the 'prompt' type of 'none' was requested but client is configured to require consent or pre-configured consent and the pre-configured consent was absent", requester.GetID(), client.GetID())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrConsentRequired)

		return nil, true
	}

	return consent, false
}

func handleOIDCAuthorizationConsentModePreConfiguredWithoutID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		config *model.OAuth2ConsentPreConfig
		err    error
	)

	if config, err = handleOIDCAuthorizationConsentModePreConfiguredGetPreConfig(ctx, client, subject, requester); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentPreConfLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if config == nil {
		if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptNone {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: the 'prompt' type of 'none' was requested but client is configured to require consent or pre-configured consent and the pre-configured consent was absent", requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrConsentRequired)

			return nil, true
		}

		return handleOIDCAuthorizationConsentGenerate(ctx, issuer, client, userSession, subject, rw, r, requester)
	}

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

	consent.PreConfiguration = sql.NullInt64{Int64: config.ID, Valid: true}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, *consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}

func handleOIDCAuthorizationConsentModePreConfiguredGetPreConfig(ctx *middlewares.AutheliaCtx, client oidc.Client, subject uuid.UUID, requester oauthelia2.Requester) (config *model.OAuth2ConsentPreConfig, err error) {
	var (
		rows *storage.ConsentPreConfigRows
	)

	ctx.Logger.Debugf(logFmtDbgConsentPreConfTryingLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), client.GetID(), subject, strings.Join(requester.GetRequestedScopes(), " "))

	if rows, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentPreConfigurations(ctx, client.GetID(), subject); err != nil {
		return nil, fmt.Errorf("error loading rows: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentPreConfRowsClose, requester.GetID(), client.GetID(), client.GetConsentPolicy(), err)
		}
	}()

	scopes, audience := requester.GetRequestedScopes(), requester.GetRequestedAudience()

	for rows.Next() {
		if config, err = rows.Get(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %w", err)
		}

		if config.HasExactGrants(scopes, audience) && config.CanConsent() {
			ctx.Logger.Debugf(logFmtDbgConsentPreConfSuccessfulLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), client.GetID(), subject, strings.Join(requester.GetRequestedScopes(), " "), config.ID)

			return config, nil
		}

		ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' using consent mode '%s' request with scopes '%s' and audience '%s' did not match pre-configured consent with scopes '%s' and audience '%s'", requester.GetID(), client.GetID(), client.GetConsentPolicy(), strings.Join(scopes, " "), strings.Join(audience, " "), strings.Join(config.Scopes, " "), strings.Join(config.Audience, " "))
	}

	ctx.Logger.Debugf(logFmtDbgConsentPreConfUnsuccessfulLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), client.GetID(), subject, strings.Join(scopes, " "), strings.Join(audience, " "))

	return nil, nil
}
