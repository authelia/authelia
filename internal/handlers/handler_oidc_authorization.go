package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
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

	requestedScopes := requester.GetRequestedScopes()
	requestedAudience := requester.GetRequestedAudience()

	isAuthInsufficient := !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel)

	if isAuthInsufficient || (isConsentMissing(userSession.OIDCWorkflowSession, requestedScopes, requestedAudience)) {
		oidcAuthorizeHandleAuthorizationOrConsentInsufficient(ctx, userSession, client, isAuthInsufficient, rw, r, requester, issuer)

		return
	}

	subject, err := ctx.Providers.OpenIDConnect.Store.GetSubject(ctx, userSession.Username)
	if err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retriving subject for user '%s': %+v", requester.GetID(), client.GetID(), userSession.Username, err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not retrieve the subject."))

		return
	}

	extraClaims := oidcGrantRequests(requester, requestedScopes, requestedAudience, &userSession)

	workflowCreated := time.Unix(userSession.OIDCWorkflowSession.CreatedTimestamp, 0)

	userSession.OIDCWorkflowSession = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the session."))

		return
	}

	if authTime, err = userSession.AuthenticatedTime(client.Policy); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred checking authentication time: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not obtain the authentication time."))

		return
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' was successfully processed, proceeding to build Authorization Response", requester.GetID(), clientID)

	oidcSession := oidc.NewSessionWithAuthorizeRequest(issuer, ctx.Providers.OpenIDConnect.KeyManager.GetActiveKeyID(),
		subject.String(), userSession.Username, extraClaims, authTime, workflowCreated, requester)

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

	ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeResponse(rw, requester, responder)
}

func oidcAuthorizeHandleAuthorizationOrConsentInsufficient(
	ctx *middlewares.AutheliaCtx, userSession session.UserSession, client *oidc.Client, isAuthInsufficient bool,
	rw http.ResponseWriter, r *http.Request,
	requester fosite.AuthorizeRequester, issuer string) {
	redirectURL := fmt.Sprintf("%s%s", issuer, string(ctx.Request.RequestURI()))

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' requires user '%s' provides consent for scopes '%s'",
		requester.GetID(), client.GetID(), userSession.Username, strings.Join(requester.GetRequestedScopes(), "', '"))

	userSession.OIDCWorkflowSession = &session.OIDCWorkflowSession{
		ClientID:                   client.GetID(),
		RequestedScopes:            requester.GetRequestedScopes(),
		RequestedAudience:          requester.GetRequestedAudience(),
		AuthURI:                    redirectURL,
		TargetURI:                  requester.GetRedirectURI().String(),
		RequiredAuthorizationLevel: client.Policy,
		CreatedTimestamp:           time.Now().Unix(),
	}

	if err := ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving session for consent: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.Fosite.WriteAuthorizeError(rw, requester, fosite.ErrServerError.WithHint("Could not save the session."))

		return
	}

	if isAuthInsufficient {
		http.Redirect(rw, r, issuer, http.StatusFound)
	} else {
		http.Redirect(rw, r, fmt.Sprintf("%s/consent", issuer), http.StatusFound)
	}
}
