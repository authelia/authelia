package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/fasthttp/router"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
)

var privateKey *rsa.PrivateKey = mustRSAKey()

func RegisterHandlers(router *router.Router, autheliaMiddleware middlewares.RequestHandlerBridge) {
	// OpenID Connect discovery: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationRequest
	router.GET("/.well-known/openid-configuration", autheliaMiddleware(WellKnownConfigurationGet))

	router.GET("/api/oidc/jwks", autheliaMiddleware(JWKsGet))
	router.GET("/api/oidc/auth", autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(AuthEndpointGet)))
	router.POST("/api/oidc/token", autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(tokenEndpoint)))

	// revoke tokens
	// http.HandleFunc("/oauth2/revoke", revokeEndpoint)
	// http.HandleFunc("/oauth2/introspect", introspectionEndpoint)
}

func NewStore() *storage.MemoryStore {
	return &storage.MemoryStore{
		IDSessions: make(map[string]fosite.Requester),
		Clients: map[string]fosite.Client{
			"authelia": &fosite.DefaultClient{
				ID:            "authelia",
				Secret:        []byte(`$2a$10$IxMdI6d.LIRZPpSfEwNoeu4rY3FhDREsxFJXikcgdRRAStxUlsuEO`), // = "foobar"
				RedirectURIs:  []string{"http://localhost:8080/oauth2/callback"},
				ResponseTypes: []string{"code"},
				GrantTypes:    []string{"implicit", "refresh_token", "authorization_code"},
				Scopes:        []string{"openid"},
			},
		},
		Users: map[string]storage.MemoryUserRelation{
			"john": {
				// This store simply checks for equality, a real storage implementation would obviously use
				// a hashing algorithm for encrypting the user password.
				Username: "john",
				Password: "secret",
			},
		},
		AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
		AccessTokens:           map[string]fosite.Requester{},
		RefreshTokens:          map[string]fosite.Requester{},
		PKCES:                  map[string]fosite.Requester{},
		AccessTokenRequestIDs:  map[string]string{},
		RefreshTokenRequestIDs: map[string]string{},
	}
}

// This is an exemplary storage instance. We will add a client and a user to it so we can use these later on.
var store = NewStore()

var config = new(compose.Config)

// Because we are using oauth2 and open connect id, we use this little helper to combine the two in one
// variable.
var start = compose.CommonStrategy{
	// alternatively you could use:
	//  OAuth2Strategy: compose.NewOAuth2JWTStrategy(mustRSAKey())
	CoreStrategy: compose.NewOAuth2HMACStrategy(config, []byte("some-super-cool-secret-that-nobody-knows"), nil),

	// open id connect strategy
	OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(config, privateKey),
}

var oauth2 = compose.Compose(
	config,
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

// A session is passed from the `/auth` to the `/token` endpoint. You probably want to store data like: "Who made the request",
// "What organization does that person belong to" and so on.
// For our use case, the session will meet the requirements imposed by JWT access tokens, HMAC access tokens and OpenID Connect
// ID Tokens plus a custom field

// newSession is a helper function for creating a new session. This may look like a lot of code but since we are
// setting up multiple strategies it is a bit longer.
// Usually, you could do:
//
//  session = new(fosite.DefaultSession)
func newSession(user string) *openid.DefaultSession {
	extra := map[string]interface{}{
		"email": fmt.Sprintf("%s@authelia.com", user),
	}

	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      "https://login.example.com:8080",
			Subject:     user,
			Audience:    []string{"https://my-client.my-application.com"},
			ExpiresAt:   time.Now().Add(time.Hour * 6),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
			Extra:       extra,
		},
		Headers: &jwt.Headers{
			Extra: make(map[string]interface{}),
		},
	}
}

func mustRSAKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	return key
}
