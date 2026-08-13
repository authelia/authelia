package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectForwardAuth(ctx AuthzContext) (object authorization.Object, err error) {
	protocol, host, uri := ctx.XForwardedProto(), ctx.XForwardedHost(), ctx.XForwardedURI()

	var (
		targetURL *url.URL
		method    []byte
	)

	if targetURL, err = getRequestURIFromForwardedHeaders(protocol, host, uri); err != nil {
		return object, fmt.Errorf("failed to get target URL: %w", err)
	}

	if method = ctx.XForwardedMethod(); len(method) == 0 {
		return object, fmt.Errorf("header 'X-Forwarded-Method' is empty")
	}

	if hasInvalidMethodCharacters(method) {
		return object, fmt.Errorf("header 'X-Forwarded-Method' with value '%s' has invalid characters", method)
	}

	return authorization.NewObjectRaw(targetURL, method), nil
}

func handleAuthzUnauthorizedForwardAuth(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
	doAuthzRedirect(ctx, authn, redirectionURL, getAuthzRedirectStatusCode(ctx, authn.Object.Method))
}
