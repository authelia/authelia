package handlers

import (
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

func handleOAuth2AuthorizationConsentModePreConfigured(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		consentID uuid.UUID
		err       error
	)

	bytesConsentID := ctx.QueryArgs().PeekBytes(qryArgConsentID)

	switch len(bytesConsentID) {
	case 0:
		return handleOAuth2AuthorizationConsentModePreConfiguredWithoutID(ctx, issuer, client, userSession, subject, rw, r, requester)
	default:
		if consentID, err = uuid.ParseBytes(bytesConsentID); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentParseChallengeID, requester.GetID(), client.GetID(), client.GetConsentPolicy(), bytesConsentID, err)

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentMalformedChallengeID)

			return nil, true
		}

		return handleOAuth2AuthorizationConsentModePreConfiguredWithID(ctx, issuer, client, userSession, subject, consentID, rw, r, requester)
	}
}

func handleOAuth2AuthorizationConsentModePreConfiguredWithID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID, consentID uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		config *model.OAuth2ConsentPreConfig
		err    error
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

	if consent.Subject.Valid && consent.Subject.UUID != uuid.Nil && consent.Subject.UUID != subject {
		ctx.Logger.Errorf(logFmtErrConsentSessionSubjectNotAuthorized, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, userSession.Username, subject, consent.Subject.UUID)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if oidc.RequesterRequiresLogin(requester, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		handleOAuth2AuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	}

	if !consent.CanGrant(ctx.Clock.Now()) {
		ctx.Logger.Errorf(logFmtErrConsentCantGrantPreConf, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return nil, true
	}

	if config, err = handleOAuth2AuthorizationConsentModePreConfiguredGetPreConfig(ctx, client, subject, requester); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentPreConfLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if config != nil {
		consent.Subject = uuid.NullUUID{UUID: subject, Valid: true}

		oidc.ConsentGrant(consent, true, config.GrantedClaims)

		consent.SetRespondedAt(ctx.Clock.Now(), config.ID)

		if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, false); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

			return nil, true
		}

		return consent, false
	}

	if !consent.IsAuthorized() {
		if consent.Responded() {
			ctx.Logger.Errorf(logFmtErrConsentCantGrantRejected, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID)

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied)

			return nil, true
		}

		handleOAuth2AuthorizationConsentRedirect(ctx, issuer, consent, client, userSession, rw, r, requester)

		return nil, true
	}

	if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptNone {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: the 'prompt' type of 'none' was requested but client is configured to require consent or pre-configured consent and the pre-configured consent was absent", requester.GetID(), client.GetID())

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oauthelia2.ErrConsentRequired)

		return nil, true
	}

	return consent, false
}

func handleOAuth2AuthorizationConsentModePreConfiguredWithoutID(ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		config *model.OAuth2ConsentPreConfig
		err    error
	)
	if config, err = handleOAuth2AuthorizationConsentModePreConfiguredGetPreConfig(ctx, client, subject, requester); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentPreConfLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return nil, true
	}

	if config == nil {
		if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptNone {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: the 'prompt' type of 'none' was requested but client is configured to require consent or pre-configured consent and the pre-configured consent was absent", requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oauthelia2.ErrConsentRequired)

			return nil, true
		}

		return handleOAuth2AuthorizationConsentGenerate(ctx, issuer, client, userSession, uuid.Nil, rw, r, requester)
	}

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

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consent.ChallengeID); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSession, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	if oidc.RequesterRequiresLogin(requester, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		handleOAuth2AuthorizationConsentPromptLoginRedirect(ctx, issuer, client, userSession, rw, r, requester, consent)

		return nil, true
	} else {
		ctx.Logger.WithFields(map[string]any{"requested_at": consent.RequestedAt, "authenticated_at": userSession.LastAuthenticatedTime(), "prompt": requester.GetRequestForm().Get("prompt")}).Debugf("Authorization Request with id '%s' on client with id '%s' is not being redirected for reauthentication", requester.GetID(), client.GetID())
	}

	oidc.ConsentGrant(consent, true, config.GrantedClaims)

	consent.SetRespondedAt(ctx.Clock.Now(), config.ID)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, false); err != nil {
		ctx.Logger.Errorf(logFmtErrConsentSaveSessionResponse, requester.GetID(), client.GetID(), client.GetConsentPolicy(), consent.ChallengeID, err)

		ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return nil, true
	}

	return consent, false
}

func handleOAuth2AuthorizationConsentModePreConfiguredGetPreConfig(ctx *middlewares.AutheliaCtx, client oidc.Client, subject uuid.UUID, requester oauthelia2.Requester) (config *model.OAuth2ConsentPreConfig, err error) {
	var (
		rows *storage.ConsentPreConfigRows
	)

	ctx.Logger.Debugf(logFmtDbgConsentPreConfTryingLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), client.GetID(), subject, strings.Join(requester.GetRequestedScopes(), " "))

	if rows, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentPreConfigurations(ctx, client.GetID(), subject, ctx.Clock.Now()); err != nil {
		return nil, fmt.Errorf("error loading rows: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			ctx.Logger.Errorf(logFmtErrConsentPreConfRowsClose, requester.GetID(), client.GetID(), client.GetConsentPolicy(), err)
		}
	}()

	var (
		requests *oidc.ClaimsRequests

		serialized, signature string
	)

	if requests, err = oidc.NewClaimRequests(requester.GetRequestForm()); err != nil {
		return nil, fmt.Errorf("error parsing claim requests: %w", err)
	} else if requests != nil {
		if serialized, signature, err = requests.Serialized(); err != nil {
			return nil, fmt.Errorf("error serializing claim requests: %w", err)
		}
	}

	scopes, audience := requester.GetRequestedScopes(), requester.GetRequestedAudience()

	log := ctx.Logger.WithFields(map[string]any{"scopes": scopes, "claims": serialized, "audience": audience, "client_id": client.GetID()})

	for rows.Next() {
		if config, err = rows.Get(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %w", err)
		}

		if !config.CanConsentAt(ctx.Clock.Now()) {
			log.Debugf("Authorization Request with id '%s' on client with id '%s' using consent mode '%s' found a matching pre-configuration with id '%d' but it is revoked, expired, or otherwise can no longer provide consent", requester.GetID(), client.GetID(), client.GetConsentPolicy(), config.ID)

			continue
		}

		if !config.HasExactGrants(scopes, audience) {
			log.Debugf("Authorization Request with id '%s' on client with id '%s' using consent mode '%s' found a matching pre-configuration with id '%d' but the configuration has scopes '%s' and audience '%s' which does not match the request", requester.GetID(), client.GetID(), client.GetConsentPolicy(), config.ID, strings.Join(config.Scopes, " "), strings.Join(config.Audience, " "))

			continue
		}

		if !config.HasClaimsSignature(signature) {
			log.Debugf("Authorization Request with id '%s' on client with id '%s' using consent mode '%s' found a matching pre-configuration with id '%d' but the configuration had the requested claims '%s' which does not match the request", requester.GetID(), client.GetID(), client.GetConsentPolicy(), config.ID, config.RequestedClaims.String)

			continue
		}

		log.Debugf(logFmtDbgConsentPreConfSuccessfulLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), client.GetID(), subject, strings.Join(requester.GetRequestedScopes(), " "), config.ID)

		return config, nil
	}

	ctx.Logger.Debugf(logFmtDbgConsentPreConfUnsuccessfulLookup, requester.GetID(), client.GetID(), client.GetConsentPolicy(), client.GetID(), subject, strings.Join(scopes, " "), strings.Join(audience, " "))

	return nil, nil
}
