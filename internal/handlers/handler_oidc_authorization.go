package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

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

	var (
		consent *model.OAuth2ConsentSession
		handled bool
	)

	if handled, consent = handleOIDCAuthorizeConsent(ctx, issuer, client, userSession, subject, rw, r, requester); handled {
		return
	}

	extraClaims := oidcGrantRequests(requester, consent, &userSession)

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

	ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeResponse(rw, requester, responder)
}

func handleOIDCAuthorizeConsent(ctx *middlewares.AutheliaCtx, rootURI string, client *oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request, requester fosite.AuthorizeRequester) (handled bool, consent *model.OAuth2ConsentSession) {
	var (
		err error
	)

	if userSession.ConsentChallengeID != nil {
		if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, *userSession.ConsentChallengeID); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred during consent session lookup: %+v", requester.GetID(), requester.GetClient().GetID(), err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Failed to lookup consent session."))

			userSession.ConsentChallengeID = nil

			if err = ctx.SaveSession(userSession); err != nil {
				ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred unlinking consent session challenge id: %+v", requester.GetID(), requester.GetClient().GetID(), err)
			}

			return true, nil
		}

		if consent.Responded() {
			userSession.ConsentChallengeID = nil

			if err = ctx.SaveSession(userSession); err != nil {
				ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving session: %+v", requester.GetID(), client.GetID(), err)

				ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the session."))

				return true, nil
			}

			if consent.Granted {
				ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: this consent session with challenge id '%s' was already granted", requester.GetID(), client.GetID(), consent.ChallengeID.String())

				ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Authorization already granted."))

				return true, nil
			}

			ctx.Logger.Debugf("Authorization Request with id '%s' loaded consent session with id '%d' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s'", requester.GetID(), consent.ID, consent.ChallengeID.String(), client.GetID(), consent.Subject.String(), strings.Join(requester.GetRequestedScopes(), " "))

			if !consent.Authorized {
				ctx.Logger.Warnf("Authorization Request with id '%s' and challenge id '%s' for client id '%s' and subject '%s' and scopes '%s' was not denied by the user durng the consent session", requester.GetID(), consent.ChallengeID.String(), client.GetID(), consent.Subject.String(), strings.Join(requester.GetRequestedScopes(), " "))

				ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrAccessDenied.WithHint("Authorization Request was denied by the user during the consent session."))

				return true, nil
			}

			return false, consent
		}
	} else {
		if consent, err = model.NewOAuth2ConsentSession(subject, requester); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred generating consent: %+v", requester.GetID(), requester.GetClient().GetID(), err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not generating the consent."))

			return true, nil
		}

		if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the consent session."))

			return true, nil
		}

		userSession.ConsentChallengeID = &consent.ChallengeID

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving user session for consent: %+v", requester.GetID(), client.GetID(), err)

			ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the user session."))

			return true, nil
		}
	}

	if client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		http.Redirect(rw, r, fmt.Sprintf("%s/consent", rootURI), http.StatusFound)
	} else {
		http.Redirect(rw, r, rootURI, http.StatusFound)
	}

	return true, consent
}
