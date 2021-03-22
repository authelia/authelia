package handlers

import (
	"github.com/fasthttp/router"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/middlewares"
)

// RegisterHandlers when provided a valid *rsa.PrivateKey and fosite.OAuth2Provider registers the handlers with the fasthttp *router.Router.
func RegisterHandlers(router *router.Router, middleware middlewares.RequestHandlerBridge, provider fosite.OAuth2Provider) {
	if provider == nil {
		// Skip Registering the handlers if the private key or Fosite has not been configured.
		return
	}

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
