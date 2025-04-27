package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

func handleAuthzPortalURLLegacy(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	if portalURL, err = handleAuthzPortalURLFromQueryLegacy(ctx); err != nil || portalURL != nil {
		return portalURL, err
	}

	return handleAuthzPortalURLFromHeader(ctx)
}

func handleAuthzPortalURLFromHeader(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	rawURL := ctx.XAutheliaURL()
	if rawURL == nil {
		return nil, nil
	}

	if portalURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return nil, err
	}

	return portalURL, nil
}

func handleAuthzPortalURLFromQuery(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	rawURL := ctx.QueryArgAutheliaURL()
	if rawURL == nil {
		return nil, nil
	}

	if portalURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return nil, err
	}

	return portalURL, nil
}

func handleAuthzPortalURLFromQueryLegacy(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	rawURL := ctx.QueryArgs().PeekBytes(qryArgRD)
	if rawURL == nil {
		return nil, nil
	}

	if portalURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return nil, err
	}

	return portalURL, nil
}

func handleAuthzAuthorizedStandard(ctx *middlewares.AutheliaCtx, headers []AuthzHeader, authn *Authn) {
	ctx.ReplyStatusCode(fasthttp.StatusOK)

	updated := ctx.Clock.Now()

	if authn.Details.GetUsername() != "" {
		for _, header := range headers {
			raw, ok := ctx.Providers.UserAttributeResolver.Resolve(header.Attribute, authn.Details, updated)

			if !ok {
				ctx.Response.Header.SetBytesK(header.Key, "")

				continue
			}

			switch value := raw.(type) {
			case string:
				ctx.Response.Header.SetBytesK(header.Key, value)
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				ctx.Response.Header.SetBytesK(header.Key, fmt.Sprintf("%d", value))
			case bool:
				ctx.Response.Header.SetBytesK(header.Key, fmt.Sprintf("%t", value))
			case []string:
				ctx.Response.Header.SetBytesK(header.Key, strings.Join(value, ","))
			default:
				ctx.Response.Header.SetBytesK(header.Key, fmt.Sprintf("%v", value))
			}
		}
	}
}

func handleAuthzAuthorizedLegacy(ctx *middlewares.AutheliaCtx, _ []AuthzHeader, authn *Authn) {
	ctx.ReplyStatusCode(fasthttp.StatusOK)

	if authn.Details.GetUsername() != "" {
		ctx.Response.Header.SetBytesK(headerRemoteUser, authn.Details.GetUsername())
		ctx.Response.Header.SetBytesK(headerRemoteGroups, strings.Join(authn.Details.GetGroups(), ","))
		ctx.Response.Header.SetBytesK(headerRemoteName, authn.Details.GetDisplayName())

		switch emails := authn.Details.GetEmails(); len(emails) {
		case 0:
			ctx.Response.Header.SetBytesK(headerRemoteEmail, "")
		default:
			ctx.Response.Header.SetBytesK(headerRemoteEmail, emails[0])
		}
	}
}
func handleAuthzUnauthorizedAuthorizationBasic(ctx *middlewares.AutheliaCtx, authn *Authn) {
	ctx.Logger.Infof("Access to '%s' is not authorized to user '%s', sending 401 response with WWW-Authenticate header requesting Basic scheme", authn.Object.URL.String(), authn.Username)

	ctx.ReplyUnauthorized()

	ctx.Response.Header.SetBytesKV(headerWWWAuthenticate, headerValueAuthenticateBasic)
}

var protoHostSeparator = []byte("://")

func getRequestURIFromForwardedHeaders(protocol, host, uri []byte) (requestURI *url.URL, err error) {
	if len(protocol) == 0 {
		return nil, fmt.Errorf("missing protocol value")
	}

	if len(host) == 0 {
		return nil, fmt.Errorf("missing host value")
	}

	value := utils.BytesJoin(protocol, protoHostSeparator, host, uri)

	if requestURI, err = url.ParseRequestURI(string(value)); err != nil {
		return nil, fmt.Errorf("failed to parse forwarded headers: %w", err)
	}

	return requestURI, nil
}

func hasInvalidMethodCharacters(v []byte) bool {
	for _, c := range v {
		if c < 0x41 || c > 0x5A {
			return true
		}
	}

	return false
}
