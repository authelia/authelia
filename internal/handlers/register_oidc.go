package handlers

import (
	"github.com/fasthttp/router"

	"github.com/authelia/authelia/internal/middlewares"
)

// RegisterOIDC registers the handlers with the fasthttp *router.Router.
func RegisterOIDC(router *router.Router, middleware middlewares.RequestHandlerBridge) {
	// TODO: Add paths for UserInfo, Flush, Logout.

	// TODO: Add OPTIONS handler.
	router.GET(oidcWellKnownPath, middleware(oidcWellKnown))

	router.GET(oidcConsentPath, middleware(oidcConsent))

	router.POST(oidcConsentPath, middleware(oidcConsentPOST))

	router.GET(oidcJWKsPath, middleware(oidcJWKs))

	router.GET(oidcAuthorizePath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcAuthorize)))

	// TODO: Add OPTIONS handler.
	router.POST(oidcTokenPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcToken)))

	router.POST(oidcIntrospectPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcIntrospect)))

	// TODO: Add OPTIONS handler.
	router.POST(oidcRevokePath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(oidcRevoke)))
}
