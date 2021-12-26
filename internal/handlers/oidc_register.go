package handlers

import (
	"github.com/fasthttp/router"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// RegisterOIDC registers the handlers with the fasthttp *router.Router. TODO: Add paths for UserInfo, Flush, Logout.
func RegisterOIDC(router *router.Router, middleware middlewares.RequestHandlerBridge) {
	// TODO: Add OPTIONS handler.
	router.GET(pathOpenIDConnectWellKnown, middleware(middlewares.CORSApplyAutomaticBasicPolicy(oidcWellKnown)))

	router.GET(pathOpenIDConnectConsent, middleware(oidcConsent))

	router.POST(pathOpenIDConnectConsent, middleware(oidcConsentPOST))

	router.GET(pathOpenIDConnectJWKs, middleware(oidcJWKs))

	router.GET(pathOpenIDConnectAuthorization, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcAuthorization)))

	// TODO: Add OPTIONS handler.
	router.POST(pathOpenIDConnectToken, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcToken)))

	router.POST(pathOpenIDConnectIntrospection, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcIntrospection)))

	router.GET(pathOpenIDConnectUserinfo, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcUserinfo)))
	router.POST(pathOpenIDConnectUserinfo, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcUserinfo)))

	// TODO: Add OPTIONS handler.
	router.POST(pathOpenIDConnectRevocation, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcRevocation)))
}
