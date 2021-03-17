package oidc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
)

// OIDCClaims represents a set of OIDC claims.
type OIDCClaims struct {
	jwt.StandardClaims

	Workflow        string   `json:"workflow"`
	Username        string   `json:"username,omitempty"`
	RequestedScopes []string `json:"requested_scopes,omitempty"`
}

func getOIDCClientConfig(clientID string, configuration schema.OpenIDConnectConfiguration) *schema.OpenIDConnectClientConfiguration {
	for _, c := range configuration.Clients {
		if clientID == c.ID {
			return &c
		}
	}

	return nil
}

// AuthorizeEndpointGet handles GET requests to the authorize endpoint.
// nolint:gocyclo
func AuthorizeEndpointGet(oauth2 fosite.OAuth2Provider) middlewares.AutheliaHandlerFunc {
	return func(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
		ctx.Logger.Debugf("Hit Authorize GET endpoint")

		ar, err := oauth2.NewAuthorizeRequest(ctx, r)
		if err != nil {
			logging.Logger().Errorf("Error occurred in NewAuthorizeRequest: %+v", err)
			oauth2.WriteAuthorizeError(rw, ar, err)

			return
		}

		clientID := ar.GetClient().GetID()

		clientConfig := getOIDCClientConfig(clientID, *ctx.Configuration.IdentityProviders.OIDC)
		if clientConfig == nil {
			err := fmt.Errorf("Unable to find related client configuration with name %s", ar.GetID())
			ctx.Logger.Error(err)
			oauth2.WriteAuthorizeError(rw, ar, err)

			return
		}

		targetURL := ar.GetRedirectURI()
		userSession := ctx.GetSession()

		// Resolve the required level of authorizations to proceed with OIDC
		requiredAuthorizationLevel := ctx.Providers.Authorizer.GetRequiredLevel(authorization.Subject{
			Username: userSession.Username,
			Groups:   userSession.Groups,
			IP:       ctx.RemoteIP(),
		}, authorization.NewObjectRaw(targetURL, []byte("GET")))

		requestedScopes := ar.GetRequestedScopes()
		requestedAudience := ar.GetRequestedAudience()

		isAuthInsufficient := !authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel, requiredAuthorizationLevel)
		isConsentMissingScopes := len(requestedScopes) > 0 && (userSession.OIDCWorkflowSession == nil ||
			utils.IsStringSlicesDifferent(requestedScopes, userSession.OIDCWorkflowSession.GrantedScopes))
		isConsentMissingAudience := len(requestedAudience) > 0 && (userSession.OIDCWorkflowSession == nil ||
			utils.IsStringSlicesDifferent(requestedAudience, userSession.OIDCWorkflowSession.GrantedAudience))

		if isAuthInsufficient || (isConsentMissingScopes || isConsentMissingAudience) {
			forwardedURI, err := ctx.GetOriginalURL()
			if err != nil {
				ctx.Logger.Errorf("%v", err)
				http.Error(rw, err.Error(), 500)

				return
			}

			ctx.Logger.Debugf("User %s must consent with scopes %s", userSession.Username, strings.Join(ar.GetRequestedScopes(), ", "))
			userSession.OIDCWorkflowSession = new(session.OIDCWorkflowSession)
			userSession.OIDCWorkflowSession.ClientID = clientID
			userSession.OIDCWorkflowSession.RequestedScopes = requestedScopes
			userSession.OIDCWorkflowSession.RequestedAudience = requestedAudience
			userSession.OIDCWorkflowSession.AuthURI = forwardedURI.String()
			userSession.OIDCWorkflowSession.TargetURI = targetURL.String()
			userSession.OIDCWorkflowSession.RequiredAuthorizationLevel = requiredAuthorizationLevel

			if err := ctx.SaveSession(userSession); err != nil {
				ctx.Logger.Errorf("%v", err)
				http.Error(rw, err.Error(), 500)

				return
			}

			uri, err := ctx.ForwardedProtoHost()
			if err != nil {
				ctx.Logger.Errorf("%v", err)
				http.Error(rw, err.Error(), 500)

				return
			}

			if isAuthInsufficient {
				http.Redirect(rw, r, uri, 302)
			} else {
				http.Redirect(rw, r, fmt.Sprintf("%s/consent", uri), 302)
			}

			return
		}

		// We grant the requested scopes at this stage.
		for _, scope := range requestedScopes {
			ar.GrantScope(scope)
		}

		for _, a := range requestedAudience {
			ar.GrantAudience(a)
		}

		userSession.OIDCWorkflowSession = nil
		if err := ctx.SaveSession(userSession); err != nil {
			ctx.Logger.Errorf("%v", err)
			http.Error(rw, err.Error(), 500)

			return
		}

		// Now that the user is authorized, we set up a session:
		oauthSession := newSession(ctx, ar.GetGrantedScopes(), ar.GetGrantedAudience())

		// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
		// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
		// to support open id connect.
		response, err := oauth2.NewAuthorizeResponse(ctx, ar, oauthSession)

		// Catch any errors, e.g.:
		// * unknown client
		// * invalid redirect
		// * ...
		if err != nil {
			ctx.Logger.Errorf("Error occurred in NewAuthorizeResponse: %+v", err)
			oauth2.WriteAuthorizeError(rw, ar, err)

			return
		}

		// TODO: Record Authorized Responses.

		oauth2.WriteAuthorizeResponse(rw, ar, response)
	}
}
