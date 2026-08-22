package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectExtAuthz(ctx AuthzContext) (object authorization.Object, err error) {
	protocol, host, uri := ctx.XForwardedProto(), ctx.Host(), ctx.AuthzPath()

	var (
		targetURL *url.URL
		method    []byte
	)

	if targetURL, err = getRequestURIFromForwardedHeaders(protocol, host, uri); err != nil {
		return object, fmt.Errorf("failed to get target URL: %w", err)
	}

	if method = ctx.Method(); len(method) == 0 {
		return object, fmt.Errorf("start line value 'Method' is empty")
	}

	if hasInvalidMethodCharacters(method) {
		return object, fmt.Errorf("start line value 'Method' with value '%s' has invalid characters", method)
	}

	return authorization.NewObjectRaw(targetURL, method), nil
}
