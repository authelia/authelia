package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OAuth2JSONWebKeySetGET returns the JSON Web Key Set. Used in OAuth 2.0 and OpenID Connect 1.0.
func OAuth2JSONWebKeySetGET(ctx *middlewares.AutheliaCtx) {
	var err error
	if _, err = ctx.IssuerURL(); err != nil {
		ctx.GetLogger().WithError(err).Errorf("JSON Web Key Set Request could not be processed: %s", oidc.ErrTextEffectiveIssuer)

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	ctx.SetContentTypeApplicationJSON()

	if err = json.NewEncoder(ctx).Encode(ctx.Providers.OpenIDConnect.Issuer.GetPublicJSONWebKeys(ctx)); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred encoding JSON web key set")
	}
}
