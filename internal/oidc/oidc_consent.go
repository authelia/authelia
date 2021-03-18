package oidc

import (
	"encoding/json"
	"fmt"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/middlewares"
)

const constAccept = "accept"
const constReject = "reject"

// ConsentGet handler serving the list consent requested by the app.
func ConsentGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.OIDCWorkflowSession == nil {
		ctx.Logger.Debug("Cannot consent when OIDC workflow has not been initiated")
		ctx.ReplyForbidden()

		return
	}

	if !authorization.IsAuthLevelSufficient(
		userSession.AuthenticationLevel,
		userSession.OIDCWorkflowSession.RequiredAuthorizationLevel) {
		ctx.Logger.Debugf("Insufficient permissions to give consent v2 %d -> %d", userSession.AuthenticationLevel, userSession.OIDCWorkflowSession.RequiredAuthorizationLevel)
		ctx.ReplyForbidden()

		return
	}

	clientConfiguration := getOIDCClientConfig(userSession.OIDCWorkflowSession.ClientID, *ctx.Configuration.IdentityProviders.OIDC)

	var body ConsentGetResponseBody
	body.Scopes = scopeNamesToScopes(userSession.OIDCWorkflowSession.RequestedScopes)
	body.Audience = audienceNamesToAudience(userSession.OIDCWorkflowSession.RequestedAudience)
	body.ClientID = userSession.OIDCWorkflowSession.ClientID

	if clientConfiguration != nil {
		body.ClientDescription = clientConfiguration.Description
	}

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Error(fmt.Errorf("Unable to set JSON body: %v", err), "Operation failed")
	}
}

// ConsentPost handler granting permissions according to the requested scopes.
func ConsentPost(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.OIDCWorkflowSession == nil {
		ctx.Logger.Debug("Cannot consent when OIDC workflow has not been initiated")
		ctx.ReplyForbidden()

		return
	}

	if !authorization.IsAuthLevelSufficient(
		userSession.AuthenticationLevel,
		userSession.OIDCWorkflowSession.RequiredAuthorizationLevel) {
		ctx.Logger.Debugf("Insufficient permissions to give consent v1 %d -> %d", userSession.AuthenticationLevel, userSession.OIDCWorkflowSession.RequiredAuthorizationLevel)
		ctx.ReplyForbidden()

		return
	}

	var body ConsentPostRequestBody
	err := json.Unmarshal(ctx.Request.Body(), &body)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to unmarshal body: %v", err), "Operation failed")
		return
	}

	if body.AcceptOrReject != constAccept && body.AcceptOrReject != constReject {
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

	if body.AcceptOrReject == constAccept {
		redirectionURL = userSession.OIDCWorkflowSession.AuthURI
		userSession.OIDCWorkflowSession.GrantedScopes = userSession.OIDCWorkflowSession.RequestedScopes
		userSession.OIDCWorkflowSession.GrantedAudience = userSession.OIDCWorkflowSession.RequestedAudience

		if err := ctx.SaveSession(userSession); err != nil {
			ctx.Error(fmt.Errorf("Unable to write session: %v", err), "Operation failed")
			return
		}
	} else if body.AcceptOrReject == constReject {
		redirectionURL = fmt.Sprintf("%s?error=access_denied&error_description=%s",
			userSession.OIDCWorkflowSession.TargetURI, "User has rejected the scopes")
		userSession.OIDCWorkflowSession = nil

		if err := ctx.SaveSession(userSession); err != nil {
			ctx.Error(fmt.Errorf("Unable to write session: %v", err), "Operation failed")
			return
		}
	}

	response := ConsentPostResponseBody{RedirectURI: redirectionURL}

	if err := ctx.SetJSONBody(response); err != nil {
		ctx.Error(fmt.Errorf("Unable to set JSON body in response"), "Operation failed")
	}
}
