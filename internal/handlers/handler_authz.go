package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// Handler is the middlewares.RequestHandler for Authz.
func (a *Authz) Handler(ctx *middlewares.AutheliaCtx) {
	var (
		object    authorization.Object
		portalURL *url.URL
		err       error
	)

	if object, err = a.fObjectGet(ctx); err != nil {
		// TODO: Adjust.
		ctx.Logger.Errorf("Error getting object: %v", err)

		ctx.ReplyUnauthorized()

		return
	}

	if !isSchemeSecure(&object.URL) {
		ctx.Logger.Errorf("Target URL '%s' has an insecure scheme '%s', only the 'https' and 'wss' schemes are supported so session cookies can be transmitted securely", object.URL.String(), object.URL.Scheme)

		ctx.ReplyUnauthorized()

		return
	}

	if portalURL, err = a.getPortalURL(ctx, &object); err != nil {
		ctx.Logger.Errorf("Target URL '%s' does not appear to be a protected domain: %+v", object.URL.String(), err)

		ctx.ReplyUnauthorized()

		return
	}

	var (
		authn         Authn
		authenticator AuthnStrategy
	)

	if authn, authenticator, err = a.authn(ctx); err != nil {
		// TODO: Adjust.
		ctx.Logger.Errorf("LOG ME: Target URL '%s' does not appear to be a protected domain: %+v", object.URL.String(), err)

		ctx.ReplyUnauthorized()

		return
	}

	authn.Object = object
	authn.Method = friendlyMethod(authn.Object.Method)

	required := ctx.Providers.Authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: authn.Details.Username,
			Groups:   authn.Details.Groups,
			IP:       ctx.RemoteIP(),
		},
		object,
	)

	switch isAuthzResult(authn.Level, required) {
	case AuthzResultForbidden:
		ctx.Logger.Infof("Access to '%s' is forbidden to user %s", object.URL.String(), authn.Username)
		ctx.ReplyForbidden()
	case AuthzResultUnauthorized:
		var handler AuthzUnauthorizedHandler

		if authenticator != nil {
			handler = authenticator.HandleUnauthorized
		} else {
			handler = a.fHandleUnauthorized
		}

		handler(ctx, &authn, a.getRedirectionURL(&object, portalURL))
	case AuthzResultAuthorized:
		a.fHandleAuthorized(ctx, &authn)
	}
}

func (a *Authz) getPortalURL(ctx *middlewares.AutheliaCtx, object *authorization.Object) (portalURL *url.URL, err error) {
	if len(a.config.Domains) == 1 {
		portalURL = a.config.Domains[0].PortalURL

		if portalURL == nil {
			rd := ctx.QueryArgs().PeekBytes(queryArgumentRedirect)
			if rd == nil {
				return nil, nil
			}

			if portalURL, err = url.ParseRequestURI(string(rd)); err != nil {
				return nil, err
			}
		}

		if portalURL != nil && strings.HasSuffix(object.Domain, a.config.Domains[0].Name) {
			return portalURL, nil
		}

		return nil, fmt.Errorf("doesn't appear to be a protected domain")
	}

	for i := 0; i < len(a.config.Domains); i++ {
		if a.config.Domains[i].Name != "" && strings.HasSuffix(object.Domain, a.config.Domains[i].Name) {
			return a.config.Domains[i].PortalURL, nil
		}
	}

	return nil, fmt.Errorf("doesn't appear to be a protected domain")
}

func (a *Authz) getRedirectionURL(object *authorization.Object, portalURL *url.URL) (redirectionURL *url.URL) {
	if portalURL == nil {
		return nil
	}

	redirectionURL, _ = url.ParseRequestURI(portalURL.String())

	qry := redirectionURL.Query()

	qry.Set(queryStrArgumentRedirect, object.URL.String())

	if object.Method != "" {
		qry.Set(queryStrArgumentRequestMethod, object.Method)
	}

	redirectionURL.RawQuery = qry.Encode()

	return redirectionURL
}

func (a *Authz) authn(ctx *middlewares.AutheliaCtx) (authn Authn, authenticator AuthnStrategy, err error) {
	for _, authenticator = range a.strategies {
		if authn, err = authenticator.Get(ctx); err != nil {
			ctx.Logger.Debugf("Error occured processing authentication: %+v", err)

			if authenticator.CanHandleUnauthorized() {
				return Authn{Type: authn.Type, Level: authentication.NotAuthenticated}, authenticator, nil
			}

			return Authn{Type: authn.Type, Level: authentication.NotAuthenticated}, nil, err
		}

		if authn.Level != authentication.NotAuthenticated {
			break
		}
	}

	if authenticator.CanHandleUnauthorized() {
		return authn, authenticator, err
	}

	return authn, nil, nil
}
