package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

//nolint:gocyclo
func oidcAuthorization(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester fosite.AuthorizeRequester
		responder fosite.AuthorizeResponder
		client    *oidc.Client
		authTime  time.Time
		issuer    string
		err       error
	)

	if requester, err = ctx.Providers.OpenIDConnect.Fosite.NewAuthorizeRequest(ctx, r); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Authorization Request failed with error: %+v", rfc)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, err)

		return
	}

	requester.GetRedirectURI()
	clientID := requester.GetClient().GetID()

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being processed", requester.GetID(), clientID)

	if client, err = ctx.Providers.OpenIDConnect.Store.GetFullClient(clientID); err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %+v", requester.GetID(), clientID, err)
		}

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, err)

		return
	}

	if issuer, err = ctx.ExternalRootURL(); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred determining issuer: %+v", requester.GetID(), clientID, err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not determine issuer."))

		return
	}

	userSession := ctx.GetSession()

	var subject uuid.UUID

	if subject, err = ctx.Providers.OpenIDConnect.Store.GetSubject(ctx, userSession.Username); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retrieving subject for user '%s': %+v", requester.GetID(), client.GetID(), userSession.Username, err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not retrieve the subject."))

		return
	}

	requestedScopes := requester.GetRequestedScopes()
	requestedAudience := requester.GetRequestedAudience()

	var (
		consent *model.OAuth2ConsentSession
		handled bool
	)

	if userSession.ConsentChallengeID != nil {
		if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, *userSession.ConsentChallengeID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retrieving oauth2 consent session with challenge id '%s' for user '%s': %+v", userSession.ConsentChallengeID.String(), requester.GetID(), client.GetID(), userSession.Username, err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not retrieve the consent session."))

			return
		}
	}

	if handled = handleOIDCAuthorizeConsent(ctx, issuer, client, userSession, consent, subject, rw, r, requester); handled {
		return
	}

	if consent.Granted {
		userSession.ConsentChallengeID = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: could not save session while rejecting already granted consent session with challenge id '%s': %+v", requester.GetID(), client.GetID(), consent.ChallengeID.String(), err)
		}

		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: this consent session with challenge id '%s' was already granted", requester.GetID(), client.GetID(), consent.ChallengeID.String())

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Authorization already granted."))

		return
	}

	extraClaims := oidcGrantRequests(requester, requestedScopes, requestedAudience, &userSession)

	if authTime, err = userSession.AuthenticatedTime(client.Policy); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred checking authentication time: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not obtain the authentication time."))

		return
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' was successfully processed, proceeding to build Authorization Response", requester.GetID(), clientID)

	oidcSession := oidc.NewSessionWithAuthorizeRequest(issuer, ctx.Providers.OpenIDConnect.KeyManager.GetActiveKeyID(),
		subject.String(), userSession.Username, extraClaims, authTime, consent, requester)

	ctx.Logger.Tracef("Authorization Request with id '%s' on client with id '%s' creating session for Authorization Response for subject '%s' with username '%s' with claims: %+v",
		requester.GetID(), oidcSession.ClientID, oidcSession.Subject, oidcSession.Username, oidcSession.Claims)
	ctx.Logger.Tracef("Authorization Request with id '%s' on client with id '%s' creating session for Authorization Response for subject '%s' with username '%s' with headers: %+v",
		requester.GetID(), oidcSession.ClientID, oidcSession.Subject, oidcSession.Username, oidcSession.Headers)

	if responder, err = ctx.Providers.OpenIDConnect.Fosite.NewAuthorizeResponse(ctx, requester, oidcSession); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Authorization Response for Request with id '%s' on client with id '%s' could not be created: %+v", requester.GetID(), clientID, rfc)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, err)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(ctx, consent.ID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the session."))

		return
	}

	userSession.ConsentChallengeID = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the session."))

		return
	}

	ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeResponse(rw, requester, responder)
}

func isOIDCConsentRequired(ctx *middlewares.AutheliaCtx, userSession *session.UserSession) (required bool, err error) {
	if userSession.ConsentChallengeID == nil {
		return true, nil
	}

	var consent *model.OAuth2ConsentSession

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, *userSession.ConsentChallengeID); err != nil {
		return false, err
	}

	return isOIDCConsentRequiredCheck(consent)
}

func isOIDCConsentRequiredCheck(consent *model.OAuth2ConsentSession) (required bool, err error) {
	if consent.Authorized {
		// TODO: Figure out of OP's are allowed to let users decide which scopes can be granted.
		// These errors can't occur under normal circumstances.
		if !utils.IsStringSliceContainsAll(consent.RequestedScopes, consent.GrantedScopes) {
			return false, errors.New("one or more requested scopes was not granted")
		}

		if !utils.IsStringSliceContainsAll(consent.RequestedAudience, consent.GrantedAudience) {
			return false, errors.New("one or more requested audiences was not granted")
		}

		return false, nil
	}

	return true, nil
}

func handleOIDCAuthorizeConsent(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	userSession session.UserSession, _ *model.OAuth2ConsentSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (handled bool) {
	var (
		required, sufficientAuth bool
		err                      error
	)

	sufficientAuth = client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel)

	if required, err = isOIDCConsentRequired(ctx, &userSession); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred checcking consent session: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not generating the consent."))

		return true
	}

	if !required {
		return false
	}

	var (
		newConsent *model.OAuth2ConsentSession
	)

	if newConsent, err = model.NewOAuth2ConsentSession(subject, requester); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: %+v", requester.GetID(), requester.GetClient().GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not generating the consent."))

		return true
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, newConsent); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the consent session."))

		return true
	}

	userSession.ConsentChallengeID = &newConsent.ChallengeID

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving user session for consent: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the user session."))

		return true
	}

	if sufficientAuth {
		http.Redirect(rw, r, fmt.Sprintf("%s/consent", rootURI), http.StatusFound)
	} else {
		http.Redirect(rw, r, rootURI, http.StatusFound)
	}

	return true
}
