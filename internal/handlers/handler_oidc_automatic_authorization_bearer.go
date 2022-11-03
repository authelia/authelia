package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func OpenIDConnectAutomaticAuthorizationBearer(ctx *middlewares.AutheliaCtx, client *oidc.Client, config *schema.OpenIDConnectAuthorizationBearerConfiguration) {
	userSession := ctx.GetSession()

	if userSession.IsAnonymous() {
		return
	}

	var (
		reqHTTPAuthorize, reqHTTPAccess *http.Request

		requester       fosite.AuthorizeRequester
		responder       fosite.AuthorizeResponder
		accessReqester  fosite.AccessRequester
		accessResponder fosite.AccessResponder
		issuer          *url.URL
		consent         *model.OAuth2ConsentSession
		subject         uuid.UUID
		authTime        time.Time
		err             error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error determining the issuer URL")

		return
	}

	state := utils.RandomString(64, utils.CharSetAlphaNumeric, true)
	nonce := utils.RandomString(64, utils.CharSetAlphaNumeric, true)
	verifier := utils.RandomString(64, utils.CharSetAlphaNumeric, true)

	challenge := sha256.New()

	challenge.Write([]byte(verifier))

	authorizeForm := url.Values{}

	authorizeForm.Set(oidc.FormScope, strings.Join(config.Scopes, " "))
	authorizeForm.Set(oidc.FormResponseType, "code")
	authorizeForm.Set(oidc.FormClientID, client.GetID())
	authorizeForm.Set(oidc.FormRedirectURI, ctx.Configuration.IdentityProviders.OIDC.AuthorizationBearers.RedirectURI.String())
	authorizeForm.Set(oidc.FormState, state)
	authorizeForm.Set(oidc.FormNonce, nonce)
	authorizeForm.Set(oidc.FormCodeChallenge, base64.RawURLEncoding.EncodeToString(challenge.Sum([]byte{})))
	authorizeForm.Set(oidc.FormCodeChallengeMethod, oidc.PKCEChallengeMethodSHA256)

	reqURLAuthorize := &url.URL{
		Scheme:   "https",
		Host:     "authelia.internal",
		Path:     oidc.EndpointPathAuthorization,
		RawQuery: authorizeForm.Encode(),
	}

	if reqHTTPAuthorize, err = http.NewRequest(http.MethodGet, reqURLAuthorize.String(), nil); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error creating the internal http request object for the authorize request")

		return
	}

	authorizeCtx := context.Background()

	if requester, err = ctx.Providers.OpenIDConnect.NewAuthorizeRequest(authorizeCtx, reqHTTPAuthorize); err != nil {
		ctx.Logger.Errorf("Automatic Authorization Bearer failed with error creating the internal authorize request: %s", fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())

		return
	}

	if subject, err = ctx.Providers.OpenIDConnect.GetSubject(authorizeCtx, client.GetSectorIdentifier(), userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error getting the subject for the user")

		return
	}

	if consent, err = model.NewOAuth2ConsentSession(subject, requester); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error generating the implicit consent")

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(authorizeCtx, *consent); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error saving the implicit consent")

		return
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(authorizeCtx, consent.ChallengeID); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error loading the implicit consent")

		return
	}

	consent.Grant()

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(authorizeCtx, *consent, false); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error saving the implicit consent response")

		return
	}

	extraClaims := oidcGrantRequests(requester, consent, &userSession)

	if authTime, err = userSession.AuthenticatedTime(client.Policy); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error retrieving authentication time")

		return
	}

	session := oidc.NewSessionWithAuthorizeRequest(issuer, ctx.Providers.OpenIDConnect.KeyManager.GetActiveKeyID(),
		userSession.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extraClaims, authTime, consent, requester)

	if responder, err = ctx.Providers.OpenIDConnect.NewAuthorizeResponse(authorizeCtx, requester, session); err != nil {
		ctx.Logger.Errorf("Automatic Authorization Bearer failed with error generating the authorize response: %s", fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())

		return
	}

	if rstate := responder.GetParameters().Get(oidc.FormState); rstate != state {
		ctx.Logger.Errorf("Automatic Authorization Bearer failed because the state did not match")

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(authorizeCtx, consent.ID); err != nil {
		ctx.Logger.WithError(err).Errorf("Automatic Authorization Bearer failed with error saving the consent session granted status")

		return
	}

	accessForm := url.Values{}

	accessForm.Set(oidc.FormGrantType, oidc.GrantTypeAuthorizationCode)
	accessForm.Set(oidc.FormCode, responder.GetParameters().Get(oidc.FormCode))
	accessForm.Set(oidc.FormCodeVerifier, verifier)
	accessForm.Set(oidc.FormClientID, client.GetID())
	accessForm.Set(oidc.FormClientSecret, config.Secret)
	accessForm.Set(oidc.FormRedirectURI, ctx.Configuration.IdentityProviders.OIDC.AuthorizationBearers.RedirectURI.String())

	reqURLAccess := &url.URL{
		Scheme: "https",
		Host:   "authelia.internal",
		Path:   oidc.EndpointPathToken,
	}

	if reqHTTPAccess, err = http.NewRequest(http.MethodPost, reqURLAccess.String(), strings.NewReader(accessForm.Encode())); err != nil {
		ctx.Logger.Errorf("Automatic Authorization Bearer failed with error creating the internal http request object for the access request")

		return
	}

	reqHTTPAccess.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	accessCtx := context.Background()
	accessSession := oidc.NewSession()

	if accessReqester, err = ctx.Providers.OpenIDConnect.NewAccessRequest(accessCtx, reqHTTPAccess, accessSession); err != nil {
		ctx.Logger.Errorf("Automatic Authorization Bearer failed with error creating the internal access request: %s", fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())

		return
	}

	if accessReqester.GetGrantTypes().ExactOne("client_credentials") {
		for _, scope := range accessReqester.GetRequestedScopes() {
			if fosite.HierarchicScopeStrategy(client.GetScopes(), scope) {
				accessReqester.GrantScope(scope)
			}
		}
	}

	if accessResponder, err = ctx.Providers.OpenIDConnect.NewAccessResponse(accessCtx, accessReqester); err != nil {
		ctx.Logger.Errorf("Automatic Authorization Bearer failed with error generating the access response: %s", fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())

		return
	}

	var (
		bearer string
		ok     bool
	)

	switch strings.ToLower(accessResponder.GetTokenType()) {
	case "bearer":
		switch config.TokenType {
		case "access_token":
			bearer = accessResponder.GetAccessToken()
		default:
			if bearer, ok = accessResponder.GetExtra("id_token").(string); ok {
				break
			}

			ctx.Logger.Errorf("Automatic Authorization Bearer failed because the access response did not contain an id token")

			return
		}
	default:
		ctx.Logger.Errorf("Automatic Authorization Bearer failed because the access response did not have a bearer token")

		return
	}

	ctx.Response.Header.SetBytesK(headerAuthorization, fmt.Sprintf("Bearer %s", bearer))
}
