package oidc

import (
	"fmt"
	"log"
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

// AuthEndpointGet handles requests to the OIDC authentication endpoint.
func AuthEndpointGet(oauth2 fosite.OAuth2Provider) middlewares.AutheliaHandlerFunc {
	return func(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
		ctx.Logger.Debugf("Hit Auth endpoint")

		// Let's create an AuthorizeRequest object!
		// It will analyze the request and extract important information like scopes, response type and others.
		ar, err := oauth2.NewAuthorizeRequest(r.Context(), r)
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

		originalURL, err := ctx.GetOriginalURL()
		if err != nil {
			err := fmt.Errorf("Unable to retrieve original URL: %v", err)
			ctx.Logger.Error(err)
			oauth2.WriteAuthorizeError(rw, ar, err)
		}

		userSession := ctx.GetSession()

		// Resolve the required level of authorizations to proceed with OIDC
		requiredAuthorizationLevel := ctx.Providers.Authorizer.GetRequiredLevel(authorization.Subject{
			Username: userSession.Username,
			Groups:   userSession.Groups,
			IP:       ctx.RemoteIP(),
		}, authorization.NewObjectRaw(originalURL, ctx.Method()))

		isAuthInsufficient := !authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel, requiredAuthorizationLevel)
		isConsentMissing := len(ar.GetRequestedScopes()) > 0 && (userSession.OIDCWorkflowSession == nil ||
			utils.IsStringSlicesDifferent(ar.GetRequestedScopes(), userSession.OIDCWorkflowSession.GrantedScopes))

		if isAuthInsufficient {
			ctx.Logger.Debugf("User %s is not sufficiently authenticated", userSession.Username)
		} else if isConsentMissing {
			ctx.Logger.Debugf("User %s must consent with scopes %s", userSession.Username, strings.Join(ar.GetRequestedScopes(), ", "))
		}

		if isAuthInsufficient || isConsentMissing {
			userSession.OIDCWorkflowSession = new(session.OIDCWorkflowSession)
			userSession.OIDCWorkflowSession.ClientID = clientID
			userSession.OIDCWorkflowSession.RequestedScopes = ar.GetRequestedScopes()
			userSession.OIDCWorkflowSession.OriginalURI = ctx.URI().String()
			userSession.OIDCWorkflowSession.RequiredAuthorizationLevel = requiredAuthorizationLevel

			if err := ctx.SaveSession(userSession); err != nil {
				ctx.Logger.Errorf("%v", err)
				http.Error(rw, err.Error(), 500)
			}

			uri, err := ctx.ForwardedProtoHost()
			if err != nil {
				ctx.Logger.Errorf("%v", err)
				http.Error(rw, err.Error(), 500)

				return
			}

			// Redirect to the authentication portal with a workflow token
			http.Redirect(rw, r, fmt.Sprintf("%s%s?workflow=openid", uri, ctx.Configuration.Server.Path), 302)

			return
		}

		scopes := ar.GetRequestedScopes()

		// We grant the requested scopes at this stage.
		for _, scope := range scopes {
			ar.GrantScope(scope)
		}

		for _, audience := range ar.GetRequestedAudience() {
			ar.GrantAudience(audience)
		}

		// Now that the user is authorized, we set up a session:
		oauthSession := newSession(ctx, scopes)

		// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
		// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
		// to support open id connect.
		response, err := oauth2.NewAuthorizeResponse(ctx, ar, oauthSession)

		// Catch any errors, e.g.:
		// * unknown client
		// * invalid redirect
		// * ...
		if err != nil {
			log.Printf("Error occurred in NewAuthorizeResponse: %+v", err)
			oauth2.WriteAuthorizeError(rw, ar, err)

			return
		}

		// TODO: Record Authorized Responses.

		oauth2.WriteAuthorizeResponse(rw, ar, response)
	}
}
