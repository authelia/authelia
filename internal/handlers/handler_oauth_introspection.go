package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/token/jwt"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OAuthIntrospectionPOST handles POST requests to the OAuth 2.0 Introspection endpoint.
//
// https://datatracker.ietf.org/doc/html/rfc7662
func OAuthIntrospectionPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		requestID uuid.UUID
		responder fosite.IntrospectionResponder
		err       error
	)

	if requestID, err = uuid.NewRandom(); err != nil {
		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, fosite.ErrServerError)

		return
	}

	oidcSession := oidc.NewSession()

	ctx.Logger.Debugf("Introspection Request with id '%s' is being processed", requestID)

	if responder, err = ctx.Providers.OpenIDConnect.NewIntrospectionRequest(ctx, req, oidcSession); err != nil {
		ctx.Logger.Errorf("Introspection Request with id '%s' failed with error: %s", requestID, oidc.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, err)

		return
	}

	ctx.Logger.Tracef("Introspection Request with id '%s' yeilded a %s (active: %t) requested at %s created with request id '%s' on client with id '%s'", requestID, responder.GetTokenUse(), responder.IsActive(), responder.GetAccessRequester().GetRequestedAt().String(), responder.GetAccessRequester().GetID(), responder.GetAccessRequester().GetClient().GetID())

	aud, introspection := oidc.IntrospectionResponseToMap(responder)

	var (
		client oidc.Client
		ok     bool
	)

	if client, ok = responder.GetAccessRequester().GetClient().(oidc.Client); !ok {
		ctx.Logger.Errorf("Introspection Request with id '%s' failed with error: %s", requestID, oidc.ErrorToDebugRFC6749Error(fosite.ErrInvalidClient.WithDebugf("The client does not implement the correct type as it's a '%T'", responder.GetAccessRequester().GetClient())))

		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, fosite.ErrInvalidClient)

		return
	}

	switch alg := client.GetIntrospectionSignedResponseAlg(); alg {
	case oidc.SigningAlgNone:
		rw.Header().Set(fasthttp.HeaderContentType, "application/json; charset=utf-8")
		rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
		rw.Header().Set(fasthttp.HeaderPragma, "no-cache")
		rw.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(rw).Encode(introspection)
	default:
		var (
			issuer *url.URL
			token  string
			jwk    *oidc.JWK
			jti    uuid.UUID
		)

		if issuer, err = ctx.IssuerURL(); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred determining issuer")

			ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, errors.WithStack(fosite.ErrServerError.WithHint("Failed to lookup required information to perform this request.").WithDebugf("The issuer could not be determined with error %+v.", err)))

			return
		}

		if jwk = ctx.Providers.OpenIDConnect.KeyManager.Get(ctx, client.GetIntrospectionSignedResponseKeyID(), alg); jwk == nil {
			ctx.Logger.WithError(err).Errorf("Introspection Request with id '%s' failed to lookup key for key manager due to likely no support for the key algorithm", requestID)

			ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, errors.WithStack(fosite.ErrServerError.WithHint("Failed to lookup required information to perform this request.").WithDebugf("The JWK matching algorithm '%s' and key id '%s' could not be found.", alg, client.GetIntrospectionSignedResponseKeyID())))

			return
		}

		if jti, err = uuid.NewRandom(); err != nil {
			ctx.Logger.WithError(err).Errorf("Introspection Request with id '%s' failed to generate a JTI", requestID)

			ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, errors.WithStack(fosite.ErrServerError.WithHint("Failed to lookup required information to perform this request.").WithDebugf("The JTI could not be generated for the Introspection JWT response type with error %+v.", err)))

			return
		}

		headers := &jwt.Headers{
			Extra: map[string]any{
				oidc.JWTHeaderKeyIdentifier: jwk.KeyID(),
				oidc.JWTHeaderKeyType:       oidc.JWTHeaderTypeValueTokenIntrospectionJWT,
			},
		}

		claims := map[string]any{
			oidc.ClaimJWTID:              jti.String(),
			oidc.ClaimIssuer:             issuer,
			oidc.ClaimIssuedAt:           time.Now().UTC().Unix(),
			oidc.ClaimTokenIntrospection: introspection,
		}

		// TODO: figure out if the aud claim should be:
		// 		- the same as the original token; or
		//      - the permitted scopes of the ORIGINAL client; or
		//      - the permitted scopes of the REQUESTING client
		if len(aud) != 0 {
			claims[oidc.ClaimAudience] = aud
		}

		if token, _, err = jwk.Strategy().Generate(ctx, claims, headers); err != nil {
			ctx.Logger.WithError(err).Errorf("Introspection Request with id '%s' failed to generate the Introspection JWT response", requestID)

			ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, errors.WithStack(fosite.ErrServerError.WithHint("Failed to generate the response.").WithDebugf("The Introspection JWT itself could not be generated with error %+v.", err)))

			return
		}

		rw.Header().Set(fasthttp.HeaderContentType, "application/token-introspection+jwt; charset=utf-8")
		rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
		rw.Header().Set(fasthttp.HeaderPragma, "no-cache")
		rw.WriteHeader(http.StatusOK)

		_, _ = rw.Write([]byte(token))
	}

	ctx.Providers.OpenIDConnect.WriteIntrospectionResponse(ctx, rw, responder)

	ctx.Logger.Debugf("Introspection Request with id '%s' was processed successfully", requestID)
}
