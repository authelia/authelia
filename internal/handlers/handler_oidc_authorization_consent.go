package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

func handleOIDCAuthorizationConsent(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		challengeID uuid.UUID
		err         error
	)

	sufficient := client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel)

	if workflowID := requester.GetRequestForm().Get("workflow_id"); workflowID != "" {
		if challengeID, err = decodeUUIDFromQueryString(workflowID); err != nil {
			handleOIDCAuthorizationConsentWithChallengeID(ctx, rootURI, client, sufficient, challengeID, rw, r, requester)
		}
	}

	return handleOIDCAuthorizationConsentOrGenerate(ctx, rootURI, client, sufficient, subject, rw, r, requester)
}

func handleOIDCAuthorizationConsentWithChallengeID(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	sufficient bool, challengeID uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		err error
	)

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, challengeID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred during consent session lookup: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Failed to lookup consent session."))

		return nil, true
	}

	if consent.Responded() {
		if consent.Granted {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: this consent session with challenge id '%s' was already granted", requester.GetID(), client.GetID(), consent.ChallengeID.String())

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Authorization already granted."))

			return nil, true
		}

		ctx.Logger.Debugf("Authorization Request with id '%s' loaded consent session with id '%d' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s'", requester.GetID(), consent.ID, consent.ChallengeID.String(), client.GetID(), consent.Subject.String(), strings.Join(requester.GetRequestedScopes(), " "))

		if consent.IsDenied() {
			ctx.Logger.Warnf("Authorization Request with id '%s' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s' was not denied by the user durng the consent session", requester.GetID(), consent.ChallengeID.String(), client.GetID(), consent.Subject.String(), strings.Join(requester.GetRequestedScopes(), " "))

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrAccessDenied)

			return nil, true
		}

		return consent, false
	}

	handleOIDCAuthorizationConsentRedirect(sufficient, rootURI, *consent, rw, r)

	return consent, true
}

func handleOIDCAuthorizationConsentOrGenerate(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	sufficient bool, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (consent *model.OAuth2ConsentSession, handled bool) {
	var (
		rows             *storage.ConsentSessionRows
		scopes, audience []string
		err              error
	)

	if rows, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionsPreConfigured(ctx, client.GetID(), subject); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' had error looking up pre-configured consent sessions: %+v", requester.GetID(), requester.GetClient().GetID(), err)
	}

	defer rows.Close()

	for rows.Next() {
		if consent, err = rows.Get(); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' had error looking up pre-configured consent sessions: %+v", requester.GetID(), requester.GetClient().GetID(), err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not lookup pre-configured consent sessions."))

			return nil, true
		}

		scopes, audience = getExpectedScopesAndAudience(requester)

		if consent.HasExactGrants(scopes, audience) && consent.CanGrant() {
			break
		}
	}

	if consent != nil && consent.HasExactGrants(scopes, audience) && consent.CanGrant() {
		return consent, false
	}

	if consent, err = model.NewOAuth2ConsentSession(subject, requester); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not generate the consent session."))

		return nil, true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, *consent); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the consent session."))

		return nil, true
	}

	handleOIDCAuthorizationConsentRedirect(sufficient, rootURI, *consent, rw, r)

	return consent, true
}

func handleOIDCAuthorizationConsentRedirect(sufficient bool, destination string, consent model.OAuth2ConsentSession, rw http.ResponseWriter, r *http.Request) {
	if sufficient {
		destination = fmt.Sprintf("%s/consent?workflow=openid_connect&workflow_id=%s", destination, encodeUUIDForQueryString(consent.ChallengeID))
	}

	http.Redirect(rw, r, destination, http.StatusFound)
}

func getExpectedScopesAndAudience(requester fosite.Requester) (scopes, audience []string) {
	audience = requester.GetRequestedAudience()
	if !utils.IsStringInSlice(requester.GetClient().GetID(), audience) {
		audience = append(audience, requester.GetClient().GetID())
	}

	return requester.GetRequestedScopes(), audience
}

func encodeUUIDForQueryString(id uuid.UUID) (value string) {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(id[:])
}

func decodeUUIDFromQueryString(value string) (id uuid.UUID, err error) {
	raw, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not decode base64 data: %w", err)
	}

	id, err = uuid.FromBytes(raw)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not parse UUID from decoded bytes: %w", err)
	}

	return id, nil
}
