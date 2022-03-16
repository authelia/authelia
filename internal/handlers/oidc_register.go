package handlers

import (
	"github.com/fasthttp/router"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// RegisterOIDC registers the handlers with the fasthttp *router.Router. TODO: Add paths for Flush, Logout.
func RegisterOIDC(router *router.Router, middleware middlewares.RequestHandlerBridge) {
	// TODO: Add OPTIONS handler.
	router.GET(oidc.WellKnownOpenIDConfigurationPath, middleware(middlewares.CORSApplyAutomaticAllowAllPolicy(wellKnownOpenIDConnectConfigurationGET)))
	router.GET(oidc.WellKnownOAuthAuthorizationServerPath, middleware(middlewares.CORSApplyAutomaticAllowAllPolicy(wellKnownOAuthAuthorizationServerGET)))

	router.GET(pathOpenIDConnectConsent, middleware(oidcConsent))

	router.POST(pathOpenIDConnectConsent, middleware(oidcConsentPOST))

	router.GET(oidc.JWKsPath, middleware(oidcJWKs))

	router.GET(oidc.AuthorizationPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcAuthorization)))
	router.GET(pathLegacyOpenIDConnectAuthorization, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcAuthorization)))

	// TODO: Add OPTIONS handler.
	router.POST(oidc.TokenPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcToken)))

	router.POST(oidc.IntrospectionPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcIntrospection)))
	router.GET(pathLegacyOpenIDConnectIntrospection, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcIntrospection)))

	router.GET(oidc.UserinfoPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcUserinfo)))
	router.POST(oidc.UserinfoPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcUserinfo)))

	// TODO: Add OPTIONS handler.
	router.POST(oidc.RevocationPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcRevocation)))
	router.POST(pathLegacyOpenIDConnectRevocation, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcRevocation)))
}
