package oidc

import (
	"time"

	"github.com/fasthttp/router"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"

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
			ResponseTypes: v.ResponseTypes,
			GrantTypes:    v.GrantTypes,
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

	// This is an exemplary storage instance. We will add a client and a user to it so we can use these later on.
	var store = NewStore(configuration)

	var oidcConfig = new(compose.Config)

	privateKey, err := utils.ParseRsaPrivateKeyFromPemStr(configuration.IssuerPrivateKey)
	if err != nil {
		panic(err)
	}

	// Because we are using oauth2 and open connect id, we use this little helper to combine the two in one
	// variable.
	/*
		var start = compose.CommonStrategy{
			CoreStrategy:               compose.NewOAuth2HMACStrategy(oidcConfig, []byte(configuration.HMACSecret), nil),
			OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(oidcConfig, privateKey),
		}


		var oauth2 = compose.Compose(
			oidcConfig,
			store,
			start,
			nil,

			// enabled handlers
			compose.OAuth2AuthorizeExplicitFactory,
			compose.OAuth2AuthorizeImplicitFactory,
			compose.OAuth2ClientCredentialsGrantFactory,
			compose.OAuth2RefreshTokenGrantFactory,
			compose.OAuth2ResourceOwnerPasswordCredentialsFactory,

			compose.OAuth2TokenRevocationFactory,
			compose.OAuth2TokenIntrospectionFactory,

			// be aware that open id connect factories need to be added after oauth2 factories to work properly.
			compose.OpenIDConnectExplicitFactory,
			compose.OpenIDConnectImplicitFactory,
			compose.OpenIDConnectHybridFactory,
			compose.OpenIDConnectRefreshFactory,
		)
	*/
	oauth2 := compose.ComposeAllEnabled(oidcConfig, store, []byte(configuration.HMACSecret), privateKey)

	// TODO: Add paths for UserInfo, Flush, Logout.

	// TODO: Add OPTIONS handler.
	router.GET(wellKnownPath, autheliaMiddleware(WellKnownConfigurationHandler))

	router.GET(consentPath, autheliaMiddleware(ConsentGet))

	router.POST(consentPath, autheliaMiddleware(ConsentPost))

	router.GET(jwksPath, autheliaMiddleware(JWKsGet(&privateKey.PublicKey)))

	router.GET(authPath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(AuthEndpointGet(oauth2))))
	router.POST(authPath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(AuthEndpointGet(oauth2))))

	// TODO: Add OPTIONS handler.
	router.POST(tokenPath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(tokenEndpoint(oauth2))))

	router.POST(introspectPath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(introspectEndpoint(oauth2))))

	// TODO: Add OPTIONS handler.
	router.POST(revokePath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(revokeEndpoint(oauth2))))
}

// A session is passed from the `/auth` to the `/token` endpoint. You probably want to store data like: "Who made the request",
// "What organization does that person belong to" and so on.
// For our use case, the session will meet the requirements imposed by JWT access tokens, HMAC access tokens and OpenID Connect
// ID Tokens plus a custom field.

// newSession is a helper function for creating a new session. This may look like a lot of code but since we are
// setting up multiple strategies it is a bit longer.
// Usually, you could do:
//
//  session = new(fosite.DefaultSession)
func newSession(ctx *middlewares.AutheliaCtx, scopes fosite.Arguments) *openid.DefaultSession {
	session := ctx.GetSession()

	extra := map[string]interface{}{}

	if len(session.Emails) != 0 && scopes.Has("email") {
		extra["email"] = session.Emails[0]
	}

	if scopes.Has("groups") {
		extra["groups"] = session.Groups
	}

	/*
		TODO: Adjust auth backends to return more profile information.
		It's probably ideal to adjust the auth providers at this time to not store 'extra' information in the session
		storage, and instead create a memory only storage for them.
		This is a simple design, have a map with a key of username, and a struct with the relevant information.
		If the
	*/
	if scopes.Has("profile") {
		extra["name"] = session.DisplayName
	}

	oidcSession := newDefaultSession(ctx)
	oidcSession.Claims.Extra = extra
	oidcSession.Claims.Subject = session.Username

	return oidcSession
}

func newDefaultSession(ctx *middlewares.AutheliaCtx) *openid.DefaultSession {
	issuer, err := ctx.ForwardedProtoHost()

	if err != nil {
		issuer = fallbackOIDCIssuer
	}

	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      issuer,
			Subject:     "",
			Audience:    []string{"https://oidc.example.com:8080"},
			ExpiresAt:   time.Now().Add(time.Hour * 6),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
			Extra:       make(map[string]interface{}),
		},
		Headers: &jwt.Headers{
			Extra: make(map[string]interface{}),
		},
	}
}
