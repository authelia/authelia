package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

func oidcConsent(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.OIDCWorkflowSession == nil {
		ctx.Logger.Debugf("Cannot consent for user %s when OIDC workflow has not been initiated", userSession.Username)
		ctx.ReplyForbidden()

		return
	}

	clientID := userSession.OIDCWorkflowSession.ClientID
	client, err := ctx.Providers.OpenIDConnect.Store.GetInternalClient(clientID)

	if err != nil {
		ctx.Logger.Debugf("Unable to find related client configuration with name '%s': %v", clientID, err)
		ctx.ReplyForbidden()

		return
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Debugf("Insufficient permissions to give consent v2 %d -> %d", userSession.AuthenticationLevel, userSession.OIDCWorkflowSession.RequiredAuthorizationLevel)
		ctx.ReplyForbidden()

		return
	}

	if err := ctx.SetJSONBody(client.GetConsentResponseBody(userSession.OIDCWorkflowSession)); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON body: %v", err), "Operation failed")
	}
}

func oidcConsentPOST(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.OIDCWorkflowSession == nil {
		ctx.Logger.Debugf("cannot consent for user %s when OIDC workflow has not been initiated", userSession.Username)
		ctx.ReplyForbidden()

		return
	}

	client, err := ctx.Providers.OpenIDConnect.Store.GetInternalClient(userSession.OIDCWorkflowSession.ClientID)

	if err != nil {
		ctx.Logger.Debugf("Unable to find related client configuration with name '%s': %v", userSession.OIDCWorkflowSession.ClientID, err)
		ctx.ReplyForbidden()

		return
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Debugf("Insufficient permissions to give consent v1 %d -> %d", userSession.AuthenticationLevel, userSession.OIDCWorkflowSession.RequiredAuthorizationLevel)
		ctx.ReplyForbidden()

		return
	}

	var body ConsentPostRequestBody
	err = json.Unmarshal(ctx.Request.Body(), &body)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to unmarshal body: %v", err), "Operation failed")
		return
	}

	if body.AcceptOrReject != accept && body.AcceptOrReject != reject {
		ctx.Logger.Infof("User %s tried to reply to consent with an unexpected verb", userSession.Username)
		ctx.ReplyBadRequest()

		return
	}

	if userSession.OIDCWorkflowSession.ClientID != body.ClientID {
		ctx.Logger.Infof("User %s consented to scopes of another client (%s) than expected (%s). Beware this can be a sign of attack",
			userSession.Username, body.ClientID, userSession.OIDCWorkflowSession.ClientID)
		ctx.ReplyBadRequest()

		return
	}

	var redirectionURL string

	if body.AcceptOrReject == accept {
		redirectionURL = userSession.OIDCWorkflowSession.AuthURI
		userSession.OIDCWorkflowSession.GrantedScopes = userSession.OIDCWorkflowSession.RequestedScopes
		userSession.OIDCWorkflowSession.GrantedAudience = userSession.OIDCWorkflowSession.RequestedAudience

		if err := ctx.SaveSession(userSession); err != nil {
			ctx.Error(fmt.Errorf("unable to write session: %v", err), "Operation failed")
			return
		}
	} else if body.AcceptOrReject == reject {
		redirectionURL = fmt.Sprintf("%s?error=access_denied&error_description=%s",
			userSession.OIDCWorkflowSession.TargetURI, "User has rejected the scopes")
		userSession.OIDCWorkflowSession = nil

		if err := ctx.SaveSession(userSession); err != nil {
			ctx.Error(fmt.Errorf("unable to write session: %v", err), "Operation failed")
			return
		}
	}

	response := ConsentPostResponseBody{RedirectURI: redirectionURL}

	if err := ctx.SetJSONBody(response); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON body in response"), "Operation failed")
	}
}
