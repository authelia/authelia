package handlers

import (
	"github.com/fasthttp/router"

	"github.com/authelia/authelia/internal/middlewares"
)

// Register when provided a non nil fosite.OAuth2Provider registers the handlers with the fasthttp *router.Router.
func Register(router *router.Router, middleware middlewares.RequestHandlerBridge) {
	// TODO: Add paths for UserInfo, Flush, Logout.
	// TODO: Add OPTIONS handler.
	router.GET(wellKnownPath, middleware(wellKnownConfigurationHandler))

	router.GET(consentPath, middleware(consentHandler))

	router.POST(consentPath, middleware(consentPOSTHandler))

	router.GET(jwksPath, middleware(jwksHandler))

	router.GET(authorizePath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(authorizeHandler)))

	// TODO: Add OPTIONS handler.
	router.POST(tokenPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(tokenHandler)))

	router.POST(introspectPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(introspectHandler)))

	// TODO: Add OPTIONS handler.
	router.POST(revokePath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(revokeHandler)))
}
