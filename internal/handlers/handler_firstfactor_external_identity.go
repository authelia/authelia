// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/externalidentity"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

var reExternalIdentityLanguage = regexp.MustCompile(`^[A-Za-z]{2,8}(-[A-Za-z0-9]{1,8}){0,7}$`)

// FirstFactorExternalIdentityProvidersGET returns the list of external identity providers for the login page.
func FirstFactorExternalIdentityProvidersGET(ctx *middlewares.AutheliaCtx) {
	all := ctx.Providers.ExternalIdentity.All()

	body := bodyGETExternalIdentityProviders{Providers: make([]bodyExternalIdentityProvider, 0, len(all))}

	for _, provider := range all {
		body.Providers = append(body.Providers, bodyExternalIdentityProvider{ID: provider.ID(), Name: provider.Name()})
	}

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred listing external identity providers: %s", errStrRespBody)
	}
}

// FirstFactorExternalIdentityPOST starts an external identity login and returns the authorization URL.
func FirstFactorExternalIdentityPOST(ctx *middlewares.AutheliaCtx) {
	var (
		id          string
		ok          bool
		redirectURI string
		err         error
	)

	if id, ok = ctx.UserValue("provider").(string); !ok {
		ctx.Logger.Error("Error occurred starting an external identity login: the provider user value wasn't set")

		ctx.SetJSONError(messageExternalIdentityLoginFailed)

		return
	}

	provider, ok := ctx.Providers.ExternalIdentity.Get(id)
	if !ok {
		ctx.Logger.WithField("provider", id).Errorf("Error occurred starting an external identity login: %s", errStrExternalIdentityProviderUnknown)

		ctx.SetJSONError(messageExternalIdentityLoginFailed)

		return
	}

	bodyJSON := bodyPOSTExternalIdentityStart{}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external identity login: %s", errStrReqBodyParse)

		ctx.SetJSONError(messageExternalIdentityLoginFailed)

		return
	}

	if redirectURI, err = externalIdentityRedirectURI(ctx, provider.ID()); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred starting an external identity login")

		ctx.SetJSONError(messageExternalIdentityLoginFailed)

		return
	}

	language := bodyJSON.Language

	// The language is only ever sent to the external provider as a hint, and it is discarded unless it is a language tag.
	if !reExternalIdentityLanguage.MatchString(language) {
		language = ""
	}

	request, err := provider.AuthorizationRequest(ctx, ctx.Providers.Random, externalidentity.AuthorizationRequestOptions{RedirectURI: redirectURI, Language: language})
	if err != nil {
		ctx.Logger.WithError(err).Error("Error occurred starting an external identity login")

		ctx.SetJSONError(messageExternalIdentityLoginFailed)

		return
	}

	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external identity login: %s", errStrUserSessionData)

		ctx.SetJSONError(messageExternalIdentityLoginFailed)

		return
	}

	userSession.ExternalIdentity = &session.ExternalIdentityFlow{
		Provider:       provider.ID(),
		State:          request.State,
		Nonce:          request.Nonce,
		CodeVerifier:   request.CodeVerifier,
		Handle:         request.Handle,
		Language:       language,
		TargetURL:      bodyJSON.TargetURL,
		RequestMethod:  bodyJSON.RequestMethod,
		KeepMeLoggedIn: bodyJSON.KeepMeLoggedIn,
		Expires:        ctx.GetClock().Now().Add(timeoutExternalIdentityFlow),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external identity login: %s", errStrUserSessionDataSave)

		ctx.SetJSONError(messageExternalIdentityLoginFailed)

		return
	}

	if err = ctx.SetJSONBody(bodyPOSTExternalIdentityStartResponse{AuthorizationURL: request.URL}); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external identity login: %s", errStrRespBody)
	}
}

func externalIdentityRedirectURI(ctx *middlewares.AutheliaCtx, id string) (uri string, err error) {
	var origin *url.URL

	if origin, err = ctx.GetOrigin(); err != nil {
		return "", fmt.Errorf("error determining the redirect uri: %w", err)
	}

	return fmt.Sprintf("%s://%s%s/api/firstfactor/external-identity/%s/callback", origin.Scheme, origin.Host, strings.TrimSuffix(ctx.BasePath(), "/"), id), nil
}

func externalIdentityRedirect(ctx *middlewares.AutheliaCtx, location string) {
	root := ctx.TemplateRootURL()

	reference, err := url.Parse(strings.TrimPrefix(location, "/"))

	if err != nil || root.Host == "" {
		ctx.SpecialRedirect(location, fasthttp.StatusFound)

		return
	}

	ctx.SpecialRedirect(root.ResolveReference(reference).String(), fasthttp.StatusFound)
}
