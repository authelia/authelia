package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// FirstFactorOpenIDConnectProvidersGET returns the list of external OpenID Connect 1.0 Providers for the login page.
func FirstFactorOpenIDConnectProvidersGET(ctx *middlewares.AutheliaCtx) {
	all := ctx.Providers.OpenIDConnectRelyingParty.All()

	body := bodyGETOpenIDConnectProviders{Providers: make([]bodyOpenIDConnectProvider, 0, len(all))}

	for _, provider := range all {
		body.Providers = append(body.Providers, bodyOpenIDConnectProvider{ID: provider.ID, Name: provider.Name})
	}

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred listing external OpenID Connect 1.0 providers: %s", errStrRespBody)
	}
}

// FirstFactorOpenIDConnectPOST starts an external OpenID Connect 1.0 login and returns the authorization URL.
func FirstFactorOpenIDConnectPOST(ctx *middlewares.AutheliaCtx) {
	var (
		id          string
		ok          bool
		redirectURI string
		err         error
	)

	if id, ok = ctx.UserValue("provider").(string); !ok {
		ctx.Logger.Error("Error occurred starting an external OpenID Connect 1.0 login: the provider user value wasn't set")

		ctx.SetJSONError(messageOpenIDConnectLoginFailed)

		return
	}

	provider, ok := ctx.Providers.OpenIDConnectRelyingParty.Get(id)
	if !ok {
		ctx.Logger.WithField("provider", id).Errorf("Error occurred starting an external OpenID Connect 1.0 login: %s", errStrOpenIDConnectProviderUnknown)

		ctx.SetJSONError(messageOpenIDConnectLoginFailed)

		return
	}

	bodyJSON := bodyPOSTOpenIDConnectStart{}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external OpenID Connect 1.0 login: %s", errStrReqBodyParse)

		ctx.SetJSONError(messageOpenIDConnectLoginFailed)

		return
	}

	if redirectURI, err = oidcRelyingPartyRedirectURI(ctx, provider.ID); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred starting an external OpenID Connect 1.0 login")

		ctx.SetJSONError(messageOpenIDConnectLoginFailed)

		return
	}

	request, err := provider.AuthorizationRequest(ctx, ctx.Providers.Random, redirectURI)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error occurred starting an external OpenID Connect 1.0 login")

		ctx.SetJSONError(messageOpenIDConnectLoginFailed)

		return
	}

	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external OpenID Connect 1.0 login: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOpenIDConnectLoginFailed)

		return
	}

	userSession.OpenIDConnect = &session.OpenIDConnectFlow{
		Provider:       provider.ID,
		State:          request.State,
		Nonce:          request.Nonce,
		CodeVerifier:   request.CodeVerifier,
		TargetURL:      bodyJSON.TargetURL,
		RequestMethod:  bodyJSON.RequestMethod,
		KeepMeLoggedIn: bodyJSON.KeepMeLoggedIn,
		Expires:        ctx.GetClock().Now().Add(timeoutOpenIDConnectFlow),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external OpenID Connect 1.0 login: %s", errStrUserSessionDataSave)

		ctx.SetJSONError(messageOpenIDConnectLoginFailed)

		return
	}

	if err = ctx.SetJSONBody(bodyPOSTOpenIDConnectStartResponse{AuthorizationURL: request.URL}); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred starting an external OpenID Connect 1.0 login: %s", errStrRespBody)
	}
}

func oidcRelyingPartyRedirectURI(ctx *middlewares.AutheliaCtx, id string) (uri string, err error) {
	var origin *url.URL

	if origin, err = ctx.GetOrigin(); err != nil {
		return "", fmt.Errorf("error determining the redirect uri: %w", err)
	}

	return fmt.Sprintf("%s://%s/api/firstfactor/openid-connect/%s/callback", origin.Scheme, origin.Host, id), nil
}
