package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func authzObjectVerifyStandard(ctx *middlewares.AutheliaCtx, object authorization.Object) (err error) {
	ctx.Logger.Debugf("Responding 200")

	if !isSchemeSecure(&object.URL) {
		return fmt.Errorf("target URL '%s' has an insecure scheme '%s', only the 'https' and 'wss' schemes are supported so session cookies can be transmitted securely", object.URL.String(), object.URL.Scheme)
	}

	if !isURLUnderProtectedDomain(&object.URL, ctx.Configuration.Session.Domain) {
		return fmt.Errorf("target URL '%s' is not on a domain which is a direct subdomain of the configured session domain '%s'", object.URL.String(), ctx.Configuration.Session.Domain)
	}

	return nil
}

func authzHandleAuthorizedStandard(ctx *middlewares.AutheliaCtx, authn *Authn) {
	ctx.ReplyStatusCode(fasthttp.StatusOK)

	if authn.Details.Username != "" {
		ctx.Response.Header.SetBytesK(headerRemoteUser, authn.Details.Username)
		ctx.Response.Header.SetBytesK(headerRemoteGroups, strings.Join(authn.Details.Groups, ","))
		ctx.Response.Header.SetBytesK(headerRemoteName, authn.Details.DisplayName)

		switch len(authn.Details.Emails) {
		case 0:
			ctx.Response.Header.SetBytesK(headerRemoteEmail, "")
		default:
			ctx.Response.Header.SetBytesK(headerRemoteEmail, authn.Details.Emails[0])
		}
	}
}

func authzHandleUnauthorizedAuthorizationBasic(ctx *middlewares.AutheliaCtx, authn *Authn, _ *url.URL) {
	ctx.Logger.Infof("Access to '%s' is not authorized to user '%s', sending 401 response with WWW-Authenticate header requesting Basic scheme", authn.Object.URL.String(), authn.Username)
	ctx.Response.Header.SetBytesKV(headerWWWAuthenticate, headerValueAuthenticateBasic)

	ctx.ReplyUnauthorized()
}
