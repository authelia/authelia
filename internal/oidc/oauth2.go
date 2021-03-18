package oidc

import (
	"github.com/fasthttp/router"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/storage"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

// NewStore creates a new OIDC store.
func NewStore(config *schema.OpenIDConnectConfiguration) *storage.MemoryStore {
	clients := make(map[string]fosite.Client)

	for _, v := range config.Clients {
		clients[v.ID] = &fosite.DefaultClient{
			ID:            v.ID,
			Secret:        []byte(v.Secret),
			RedirectURIs:  v.RedirectURIs,
			GrantTypes:    v.GrantTypes,
			ResponseTypes: v.ResponseTypes,
			Scopes:        v.Scopes,
		}
	}

	return &storage.MemoryStore{
		IDSessions:             make(map[string]fosite.Requester),
		Clients:                clients,
		Users:                  map[string]storage.MemoryUserRelation{},
		AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
		AccessTokens:           map[string]fosite.Requester{},
		RefreshTokens:          map[string]storage.StoreRefreshToken{},
		PKCES:                  map[string]fosite.Requester{},
		AccessTokenRequestIDs:  map[string]string{},
		RefreshTokenRequestIDs: map[string]string{},
	}
}

// InitializeOIDC configures the fasthttp router to provide OIDC.
func InitializeOIDC(configuration *schema.OpenIDConnectConfiguration, router *router.Router, autheliaMiddleware middlewares.RequestHandlerBridge) {
	if configuration == nil {
		return
	}

	// TODO: Replace this storage interface with a persistent type.
	var store = NewStore(configuration)

	var oidcConfig = new(compose.Config)

	privateKey, err := utils.ParseRsaPrivateKeyFromPemStr(configuration.IssuerPrivateKey)
	if err != nil {
		panic(err)
	}

	oauth2 := compose.ComposeAllEnabled(oidcConfig, store, []byte(configuration.HMACSecret), privateKey)

	// TODO: Add paths for UserInfo, Flush, Logout.

	// TODO: Add OPTIONS handler.
	router.GET(wellKnownPath, autheliaMiddleware(WellKnownConfigurationHandler))

	router.GET(consentPath, autheliaMiddleware(ConsentGet))

	router.POST(consentPath, autheliaMiddleware(ConsentPost))

	router.GET(jwksPath, autheliaMiddleware(JWKsGet(&privateKey.PublicKey)))

	router.GET(authorizePath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(AuthorizeEndpoint(oauth2))))

	// TODO: Add OPTIONS handler.
	router.POST(tokenPath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(tokenEndpoint(oauth2))))

	router.POST(introspectPath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(introspectEndpoint(oauth2))))

	// TODO: Add OPTIONS handler.
	router.POST(revokePath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(revokeEndpoint(oauth2))))
}
