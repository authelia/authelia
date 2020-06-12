package oidc

import (
	"encoding/json"
	"fmt"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/middlewares"
)

// ConsentPostRequestBody schema of the request body of the consent POST endpoint.
type ConsentPostRequestBody struct {
	ClientID string `json:"client_id"`
}

// ConsentGetResponseBody schema of the response body of the consent GET endpoint.
type ConsentGetResponseBody struct {
	ClientID string   `json:"client_id"`
	Scopes   []string `json:"scopes"`
}

// ConsentGet handler serving the list consent requested by the app.
func ConsentGet(req *middlewares.AutheliaCtx) {
	userSession := req.GetSession()

	if userSession.OIDCWorkflowSession == nil {
		req.Logger.Debug("Cannot consent when OIDC workflow has not been initiated")
		req.ReplyForbidden()

		return
	}

	if authorization.IsAuthLevelSufficient(
		userSession.AuthenticationLevel,
		userSession.OIDCWorkflowSession.RequiredAuthorizationLevel) {
		req.Logger.Debug("Insufficient permissions to give consent")
		req.ReplyForbidden()

		return
	}

	var body ConsentGetResponseBody
	body.Scopes = userSession.OIDCWorkflowSession.RequestedScopes
	body.ClientID = userSession.OIDCWorkflowSession.ClientID

	if err := req.SetJSONBody(body); err != nil {
		req.Error(fmt.Errorf("Unable to set JSON body: %v", err), "Operation failed")
	}
}

// ConsentPost handler granting permissions according to the requested scopes.
func ConsentPost(req *middlewares.AutheliaCtx) {
	userSession := req.GetSession()

	if userSession.OIDCWorkflowSession == nil {
		req.Logger.Debug("Cannot consent when OIDC workflow has not been initiated")
		req.ReplyForbidden()

		return
	}

	if authorization.IsAuthLevelSufficient(
		userSession.AuthenticationLevel,
		userSession.OIDCWorkflowSession.RequiredAuthorizationLevel) {
		req.Logger.Debug("Insufficient permissions to give consent")
		req.ReplyForbidden()

		return
	}

	var body ConsentPostRequestBody
	err := json.Unmarshal(req.Request.Body(), &body)

	if err != nil {
		req.Error(fmt.Errorf("Unable to unmarshal body: %v", err), "Operation failed")
		return
	}

	if userSession.OIDCWorkflowSession.ClientID != body.ClientID {
		req.Logger.Infof("User %s consented to scopes of another client (%s) than expected (%s). Beware this can be a sign of attack",
			userSession.Username, body.ClientID, userSession.OIDCWorkflowSession.ClientID)
		req.ReplyBadRequest()

		return
	}

	userSession.OIDCWorkflowSession.GrantedScopes = userSession.OIDCWorkflowSession.RequestedScopes
	if err := req.SaveSession(userSession); err != nil {
		req.Error(fmt.Errorf("Unable to write session: %v", err), "Operation failed")
		return
	}

	req.Redirect(userSession.OIDCWorkflowSession.OriginalURI, 302)
}
