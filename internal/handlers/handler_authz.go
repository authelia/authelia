package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// Handler is the middlewares.RequestHandler for Authz.
func (authz *Authz) Handler(ctx *middlewares.AutheliaCtx) {
	var (
		object      authorization.Object
		autheliaURL *url.URL
		err         error
	)

	if object, err = authz.handleGetObject(ctx); err != nil {
		// TODO: Adjust.
		ctx.Logger.Errorf("Error getting object: %v", err)

		ctx.ReplyUnauthorized()

		return
	}

	if !utils.IsURISecure(object.URL) {
		ctx.Logger.Errorf("Target URL '%s' has an insecure scheme '%s', only the 'https' and 'wss' schemes are supported so session cookies can be transmitted securely", object.URL.String(), object.URL.Scheme)

		ctx.ReplyUnauthorized()

		return
	}

	var provider *session.Session

	if provider, err = ctx.GetSessionProviderByTargetURL(object.URL); err != nil {
		ctx.Logger.Errorf("Target URL '%s' does not appear to be a protected domain: %+v", object.URL.String(), err)

		ctx.ReplyUnauthorized()

		return
	}

	if autheliaURL, err = authz.getAutheliaURL(ctx, provider); err != nil {
		ctx.Logger.Errorf("Target URL '%s' does not appear to be a protected domain: %+v", object.URL.String(), err)

		ctx.ReplyUnauthorized()

		return
	}

	var (
		authn         Authn
		authenticator AuthnStrategy
	)

	if authn, authenticator, err = authz.authn(ctx, provider); err != nil {
		// TODO: Adjust.
		ctx.Logger.Errorf("LOG ME: Target URL '%s' does not appear to be a protected domain: %+v", object.URL.String(), err)

		ctx.ReplyUnauthorized()

		return
	}

	authn.Object = object
	authn.Method = friendlyMethod(authn.Object.Method)

	ruleHasSubject, required := ctx.Providers.Authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: authn.Details.Username,
			Groups:   authn.Details.Groups,
			IP:       ctx.RemoteIP(),
		},
		object,
	)

	switch isAuthzResult(authn.Level, required, ruleHasSubject) {
	case AuthzResultForbidden:
		ctx.Logger.Infof("Access to '%s' is forbidden to user %s", object.URL.String(), authn.Username)
		ctx.ReplyForbidden()
	case AuthzResultUnauthorized:
		var handler HandlerAuthzUnauthorized

		if authenticator != nil {
			handler = authenticator.HandleUnauthorized
		} else {
			handler = authz.handleUnauthorized
		}

		handler(ctx, &authn, authz.getRedirectionURL(&object, autheliaURL))
	case AuthzResultAuthorized:
		authz.handleAuthorized(ctx, &authn)
	}
}

func (authz *Authz) getAutheliaURL(ctx *middlewares.AutheliaCtx, provider *session.Session) (autheliaURL *url.URL, err error) {
	if authz.handleGetAutheliaURL == nil {
		return nil, nil
	}

	if autheliaURL, err = authz.handleGetAutheliaURL(ctx); err != nil {
		return nil, err
	}

	if autheliaURL != nil {
		return autheliaURL, nil
	}

	if provider.Config.AutheliaURL != nil {
		if authz.legacy {
			return nil, nil
		}

		return provider.Config.AutheliaURL, nil
	}

	return nil, fmt.Errorf("authelia url lookup failed")
}

func (authz *Authz) getRedirectionURL(object *authorization.Object, autheliaURL *url.URL) (redirectionURL *url.URL) {
	if autheliaURL == nil {
		return nil
	}

	redirectionURL, _ = url.ParseRequestURI(autheliaURL.String())

	qry := redirectionURL.Query()

	qry.Set(queryArgRD, object.URL.String())

	if object.Method != "" {
		qry.Set(queryArgRM, object.Method)
	}

	redirectionURL.RawQuery = qry.Encode()

	return redirectionURL
}

func (authz *Authz) authn(ctx *middlewares.AutheliaCtx, provider *session.Session) (authn Authn, strategy AuthnStrategy, err error) {
	for _, strategy = range authz.strategies {
		if authn, err = strategy.Get(ctx, provider); err != nil {
			if strategy.CanHandleUnauthorized() {
				return Authn{Type: authn.Type, Level: authentication.NotAuthenticated}, strategy, nil
			}

			return Authn{Type: authn.Type, Level: authentication.NotAuthenticated}, nil, err
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
