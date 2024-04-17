package handlers

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

// OpenIDConnectConsentGET returns all OpenIDConnect clients.
func OpenIDConnectClientConfigGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading OIDC Clients: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Errorf("Error occurred loading OIDC Clients")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	clients := convertToOpenIDConnectClients(ctx.Configuration.IdentityProviders.OIDC.Clients)

	if err = ctx.SetJSONBody(clients); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred fetching OIDC Clients from OpenIDConnectProvider: %s", errStrRespBody)
	}
}

// convertToOpenIDConnectClients converts all clients loaded from config to sanitized model.OpenIDConnectClient.
func convertToOpenIDConnectClients(clients []schema.IdentityProvidersOpenIDConnectClient) []model.OpenIDConnectClient {
	var convertedClients = make([]model.OpenIDConnectClient, 0, len(clients))

	for _, c := range clients {
		convertedClient := convertSingleOpenIDConnectClient(c)

		convertedClients = append(convertedClients, convertedClient)
	}

	return convertedClients
}

// converSingleOpenIDConnectClient maps the schema.IdentityProvidersOpenIDConnectClient from Configuration struct to sanitized model.OpenIDConnectClient.
func convertSingleOpenIDConnectClient(c schema.IdentityProvidersOpenIDConnectClient) model.OpenIDConnectClient {
	convertedClient := model.OpenIDConnectClient{
		ID:                  c.ID,
		Name:                c.Name,
		SectorIdentifierURI: c.SectorIdentifierURI,
		Public:              c.Public,
		RedirectURIs:        c.RedirectURIs,
		RequestURIs:         c.RequestURIs,
		Audience:            c.Audience,
		Scopes:              c.Scopes,
	}

	return convertedClient
}

// TODO (Crowley723): handle loading of OIDC Client config from database - requires configuration in database.
