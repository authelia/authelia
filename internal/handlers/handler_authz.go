package handlers

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

type AuthzContext interface {
	context.Context

	GetLogger() *logrus.Entry
	GetConfiguration() schema.Configuration
	GetClock() clock.Provider
	GetProviders() middlewares.Providers
	GetUserProvider() authentication.UserProvider
	GetRandom() (random random.Provider)
	GetProviderUserAttributeResolver() expression.UserAttributeResolver
	GetJWTWithTimeFuncOption() (option jwt.ParserOption)

	Method() (method []byte)
	Host() (host []byte)
	XForwardedMethod() (method []byte)
	XForwardedProto() (proto []byte)
	XForwardedHost() (host []byte)
	XForwardedURI() (uri []byte)
	XOriginalMethod() (method []byte)
	XOriginalURL() (uri []byte)
	GetXOriginalURLOrXForwardedURL() (requestURI *url.URL, err error)
	XAutheliaURL() []byte
	QueryArgAutheliaURL() []byte
	RootURL() (issuerURL *url.URL)
	IssuerURL() (issuerURL *url.URL, err error)

	GetSessionManagerByTargetURI(targetURL *url.URL) (manager session.Manager, err error)

	GetRequestQueryArgValue(key []byte) (value []byte)
	GetRequestHeaderValue(key []byte) (value []byte)
	SetResponseHeaderValue(key []byte, value string)
	SetResponseHeaderValueBytes(key, value []byte)

	AuthzPath() (uri []byte)
	IsXHR() (xhr bool)
	AcceptsMIME(mime string) (ok bool)

	ReplyStatusCode(statusCode int)
	ReplyUnauthorized()
	ReplyForbidden()
	SpecialRedirect(uri string, statusCode int)
	SpecialRedirectNoBody(uri string, statusCode int)

	RecordAuthn(success, banned bool, authType string)
	RemoteIP() net.IP
	GetSessionProviderByTargetURI(targetURL *url.URL) (provider *session.Session, err error)
}

// Handler is the middlewares.RequestHandler for Authz.
func (authz *Authz) Handler(ctx AuthzContext) {
	var (
		object      authorization.Object
		autheliaURL *url.URL
		manager     session.Manager
		err         error
	)

	if object, err = authz.handleGetObject(ctx); err != nil {
		ctx.GetLogger().WithError(err).Error("Error getting Target URL and Request Method")

		ctx.ReplyStatusCode(authz.config.StatusCodeBadRequest)

		return
	}

	if !utils.IsURISecure(object.URL) {
		ctx.GetLogger().Errorf("Target URL '%s' has an insecure scheme '%s', only the 'https' and 'wss' schemes are supported so session cookies can be transmitted securely", object.URL.String(), object.URL.Scheme)

		ctx.ReplyStatusCode(authz.config.StatusCodeBadRequest)

		return
	}

	if manager, err = ctx.GetSessionManagerByTargetURI(object.URL); err != nil || manager.GetSessionConfig().Domain == "" {
		ctx.GetLogger().WithError(err).WithField("target_url", object.URL.String()).Error("Target URL does not appear to have a relevant session cookies configuration")

		ctx.ReplyStatusCode(authz.config.StatusCodeBadRequest)

		return
	}

	if autheliaURL, err = authz.getAutheliaURL(ctx, manager); err != nil {
		ctx.GetLogger().WithError(err).WithField("target_url", object.URL.String()).Error("Error occurred trying to determine the external Authelia URL for Target URL")

		ctx.ReplyStatusCode(authz.config.StatusCodeBadRequest)

		return
	}

	var (
		authn    *Authn
		strategy AuthnStrategy
	)

	authn, strategy, err = authz.authn(ctx, manager, &object)

	authn.Object = object
	authn.Method = friendlyMethod(authn.Object.Method)

	ruleHasSubject, required := ctx.GetProviders().Authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: authn.Details.Username,
			Groups:   authn.Details.Groups,
			ClientID: authn.ClientID,
			IP:       ctx.RemoteIP(),
		},
		object,
	)

	if err != nil {
		authn.Object = object

		if !ruleHasSubject && required != authorization.Bypass {
			switch {
			case strategy == nil:
				ctx.ReplyUnauthorized()
			case strategy.HeaderStrategy():
				ctx.GetLogger().WithError(err).Error("Error occurred while attempting to authenticate a request")

				strategy.HandleUnauthorized(ctx, authn, authz.getRedirectionURL(&object, autheliaURL))

				return
			}
		}

		ctx.GetLogger().WithError(err).Debug("Error occurred while attempting to authenticate a request but the matched rule was a bypass rule")
	}

	switch isAuthzResult(authn.Level, required, ruleHasSubject) {
	case AuthzResultForbidden:
		ctx.GetLogger().Infof("Access to '%s' is forbidden to user '%s'", object.URL.String(), authn.Username)
		ctx.ReplyForbidden()
	case AuthzResultUnauthorized:
		var handler HandlerAuthzUnauthorized

		if strategy != nil {
			handler = strategy.HandleUnauthorized
		} else {
			handler = authz.handleUnauthorized
		}

		handler(ctx, authn, authz.getRedirectionURL(&object, autheliaURL))
	case AuthzResultAuthorized:
		authz.handleAuthorized(ctx, authn)
	}
}

func (authz *Authz) getAutheliaURL(ctx AuthzContext, manager session.Manager) (autheliaURL *url.URL, err error) {
	if autheliaURL, err = authz.handleGetAutheliaURL(ctx); err != nil {
		return nil, err
	}

	config := manager.GetSessionConfig()

	switch {
	case authz.implementation == AuthzImplLegacy:
		return autheliaURL, nil
	case autheliaURL != nil:
		switch {
		case utils.HasURIDomainSuffix(autheliaURL, config.Domain):
			return autheliaURL, nil
		default:
			return nil, fmt.Errorf("authelia url '%s' is not valid for detected domain '%s' as the url does not have the domain as a suffix", autheliaURL.String(), config.Domain)
		}
	}

	if config.AutheliaURL != nil {
		return config.AutheliaURL, nil
	}

	return nil, fmt.Errorf("authelia url lookup failed")
}

func (authz *Authz) getRedirectionURL(object *authorization.Object, autheliaURL *url.URL) (redirectionURL *url.URL) {
	if autheliaURL == nil {
		return nil
	}

	redirectionURL, _ = url.ParseRequestURI(autheliaURL.String())

	if redirectionURL.Path == "" {
		redirectionURL.Path = "/"
	}

	qry := redirectionURL.Query()

	qry.Set(queryArgRD, object.URL.String())

	if object.Method != "" {
		qry.Set(queryArgRM, object.Method)
	}

	redirectionURL.RawQuery = qry.Encode()

	return redirectionURL
}

func (authz *Authz) authn(ctx AuthzContext, manager session.Manager, object *authorization.Object) (authn *Authn, strategy AuthnStrategy, err error) {
	for _, strategy = range authz.strategies {
		if authn, err = strategy.Get(ctx, manager, object); err != nil {
			// Ensure an error returned can never result in an authenticated user.
			authn.Level = authentication.NotAuthenticated
			authn.Username = anonymous
			authn.ClientID = ""
			authn.Details = authentication.UserDetails{}

			if strategy.CanHandleUnauthorized() {
				return authn, strategy, err
			}

			return authn, nil, err
		}

		if authn.Level != authentication.NotAuthenticated {
			break
		}
	}

	if strategy.CanHandleUnauthorized() {
		return authn, strategy, err
	}

	return authn, nil, nil
}
