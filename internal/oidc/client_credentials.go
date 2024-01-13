package oidc

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/plaintext"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/pkce"
	"github.com/ory/x/errorsx"
	"github.com/valyala/fasthttp"
	"gopkg.in/square/go-jose.v2"
)

// NewHasher returns a new Hasher.
func NewHasher() (hasher *Hasher, err error) {
	hasher = &Hasher{}

	if hasher.decoder, err = crypt.NewDefaultDecoder(); err != nil {
		return nil, err
	}

	if err = plaintext.RegisterDecoderPlainText(hasher.decoder); err != nil {
		return nil, err
	}

	return hasher, nil
}

// Hasher implements the fosite.Hasher interface and adaptively compares hashes.
type Hasher struct {
	decoder algorithm.DecoderRegister
}

// Compare compares the hash with the data and returns an error if they don't match.
func (h Hasher) Compare(_ context.Context, hash, data []byte) (err error) {
	var digest algorithm.Digest

	if digest, err = h.decoder.Decode(string(hash)); err != nil {
		return err
	}

	if digest.MatchBytes(data) {
		return nil
	}

	return errPasswordsDoNotMatch
}

// Hash creates a new hash from data.
func (h Hasher) Hash(_ context.Context, data []byte) (hash []byte, err error) {
	return data, nil
}

// DefaultClientAuthenticationStrategy is a copy of fosite's with the addition of the client_secret_jwt method and some
// minor superficial changes.
func (p *OpenIDConnectProvider) DefaultClientAuthenticationStrategy(ctx context.Context, r *http.Request, form url.Values) (c fosite.Client, err error) {
	switch assertionType := form.Get(FormParameterClientAssertionType); assertionType {
	case "":
		break
	case ClientAssertionJWTBearerType:
		return p.JWTBearerClientAuthenticationStrategy(ctx, r, form)
	default:
		return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("Unknown client_assertion_type '%s'.", assertionType))
	}

	clientID, clientSecret, method, err := clientCredentialsFromRequest(r.Header, form)
	if err != nil {
		return nil, err
	}

	var client Client

	if client, err = p.Store.GetFullClient(ctx, clientID); err != nil {
		if errors.Is(err, fosite.ErrInvalidClient) {
			return nil, err
		}

		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	if fclient, ok := client.(*FullClient); ok {
		cmethod := fclient.GetTokenEndpointAuthMethod()

		switch {
		case method != cmethod:
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf(errHintFmtClientAuthMethodMismatch, cmethod, method, method))
		case method != ClientAuthMethodNone && fclient.IsPublic():
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("The OAuth 2.0 Client is not a confidential client however the client authentication method '%s' was used which is not permitted as it's only permitted on confidential clients.", method))
		case method == ClientAuthMethodNone && !fclient.IsPublic():
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("The OAuth 2.0 Client is a confidential client however the client authentication method '%s' was used which is not permitted as it's not permitted on confidential clients.", method))
		}
	}

	if client.IsPublic() {
		return client, nil
	}

	if err = p.checkClientSecret(ctx, client, []byte(clientSecret)); err != nil {
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	return client, nil
}

//nolint:gocyclo
func (p *OpenIDConnectProvider) JWTBearerClientAuthenticationStrategy(ctx context.Context, r *http.Request, form url.Values) (client fosite.Client, err error) {
	if form.Has(FormParameterClientSecret) {
		return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("The client_secret request parameter must not be set when using client_assertion_type of '%s'.", ClientAssertionJWTBearerType))
	}

	if value := r.Header.Get(fasthttp.HeaderAuthorization); len(value) != 0 {
		return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("The Authorization request header must not be set when using client_assertion_type of '%s'.", ClientAssertionJWTBearerType))
	}

	var claims jwt.MapClaims

	client, _, claims, err = p.parseJWTAssertion(ctx, form)

	switch {
	case err == nil:
		break
	default:
		var e *fosite.RFC6749Error

		if errors.As(err, &e) {
			return nil, e
		}

		rfc := fosite.ErrInvalidClient.
			WithHint("Unable to verify the integrity of the 'client_assertion' value.").
			WithWrap(err)

		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, errorsx.WithStack(rfc.WithDebug("The token is malformed."))
		case errors.Is(err, jwt.ErrTokenUsedBeforeIssued):
			return nil, errorsx.WithStack(rfc.WithDebug("The token was used before it was issued."))
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, errorsx.WithStack(rfc.WithDebug("The token is expired."))
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, errorsx.WithStack(rfc.WithDebug("The token isn't valid yet."))
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, errorsx.WithStack(rfc.WithDebug("The signature is invalid."))
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			return nil, errorsx.WithStack(rfc.WithDebug("The token claims are invalid."))
		default:
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Unable to verify the integrity of the 'client_assertion' value.").WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
		}
	}

	var (
		iss, sub, jti string
		ok            bool
		exp           *jwt.NumericDate
	)

	if iss, err = claims.GetIssuer(); err != nil {
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'iss' from 'client_assertion' is invalid.").WithWrap(err))
	}

	if sub, err = claims.GetSubject(); err != nil {
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'sub' from 'client_assertion' is invalid.").WithWrap(err))
	}

	if jti, err = JTIFromMapClaims(claims); err != nil {
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'jti' from 'client_assertion' is invalid.").WithWrap(err))
	}

	tokenURL := p.Config.GetTokenURL(ctx)

	switch {
	case iss != client.GetID():
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'iss' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client.").WithDebugf(errDebugFmtParameterMatchClaim, ClaimIssuer, iss, FormParameterClientID, client.GetID()))
	case sub != client.GetID():
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'sub' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client.").WithDebugf(errDebugFmtParameterMatchClaim, ClaimSubject, sub, FormParameterClientID, client.GetID()))
	case tokenURL == "":
		return nil, errorsx.WithStack(fosite.ErrMisconfiguration.WithHint("The authorization server's token endpoint URL has not been set."))
	case len(jti) == 0:
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'jti' from 'client_assertion' must be set but it is not."))
	case p.Store.ClientAssertionJWTValid(ctx, jti) != nil:
		return nil, errorsx.WithStack(fosite.ErrJTIKnown.WithHint("Claim 'jti' from 'client_assertion' MUST only be used once."))
	}

	if exp, err = claims.GetExpirationTime(); err != nil {
		var epoch int64

		if epoch, ok = claims[ClaimExpirationTime].(int64); ok {
			exp = jwt.NewNumericDate(time.Unix(epoch, 0))
		} else {
			return nil, errorsx.WithStack(err)
		}
	}

	if err = p.Store.SetClientAssertionJWT(ctx, jti, exp.Time); err != nil {
		return nil, err
	}

	var (
		found bool
	)

	if auds, ok := claims[ClaimAudience].([]any); ok {
		for _, aud := range auds {
			if audience, ok := aud.(string); ok && audience == tokenURL {
				found = true
				break
			}
		}
	}

	if !found {
		return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("Claim 'aud' from 'client_assertion' must match the authorization server's token endpoint '%s'.", tokenURL))
	}

	return client, nil
}

//nolint:gocyclo
func (p *OpenIDConnectProvider) parseJWTAssertion(ctx context.Context, form url.Values) (client fosite.Client, token *jwt.Token, claims jwt.MapClaims, err error) {
	assertion := form.Get(FormParameterClientAssertion)
	if len(assertion) == 0 {
		return nil, nil, nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("The 'client_assertion' request parameter must be set when using 'client_assertion_type' of '%s'.", ClientAssertionJWTBearerType))
	}

	var (
		clientID string
	)

	clientID = form.Get(FormParameterClientID)

	parserOpts := []jwt.ParserOption{
		jwt.WithIssuedAt(),
		jwt.WithStrictDecoding(),
	}

	if octx, ok := ctx.(Context); ok {
		parserOpts = append(parserOpts, octx.GetJWTWithTimeFuncOption())
	}

	token, err = jwt.ParseWithClaims(assertion, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		var (
			ok bool
		)

		claims, ok = token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("The claims could not be parsed due to an unknown error."))
		}

		if clientID, err = parseJWTAssertionClientID(clientID, claims); err != nil {
			return nil, err
		}

		if client, err = p.Store.GetFullClient(ctx, clientID); err != nil {
			if errors.Is(err, fosite.ErrInvalidClient) {
				return nil, err
			}

			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
		}

		fclient, ok := client.(*FullClient)
		if !ok {
			return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHint("The client configuration does not support OpenID Connect specific authentication methods."))
		}

		switch fclient.GetTokenEndpointAuthMethod() {
		case ClientAuthMethodPrivateKeyJWT:
			switch token.Method {
			case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512,
				jwt.SigningMethodPS256, jwt.SigningMethodPS384, jwt.SigningMethodPS512,
				jwt.SigningMethodES256, jwt.SigningMethodES384, jwt.SigningMethodES512:
				break
			default:
				return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("This requested OAuth 2.0 client supports client authentication method '%s', however the '%s' JWA is not supported with this method.", ClientAuthMethodPrivateKeyJWT, token.Header[JWTHeaderKeyAlgorithm]))
			}
		case ClientAuthMethodClientSecretJWT:
			switch token.Method {
			case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
				break
			default:
				return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("This requested OAuth 2.0 client supports client authentication method '%s', however the '%s' JWA is not supported with this method.", ClientAuthMethodClientSecretJWT, token.Header[JWTHeaderKeyAlgorithm]))
			}
		case ClientAuthMethodNone:
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("This requested OAuth 2.0 client does not support client authentication, however 'client_assertion' was provided in the request."))
		case ClientAuthMethodClientSecretPost:
			fallthrough
		case ClientAuthMethodClientSecretBasic:
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("This requested OAuth 2.0 client only supports client authentication method '%s', however 'client_assertion' was provided in the request.", fclient.GetTokenEndpointAuthMethod()))
		default:
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("This requested OAuth 2.0 client only supports client authentication method '%s', however that method is not supported by this server.", fclient.GetTokenEndpointAuthMethod()))
		}

		if fclient.GetTokenEndpointAuthSigningAlgorithm() != fmt.Sprintf("%s", token.Header[JWTHeaderKeyAlgorithm]) {
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("The 'client_assertion' uses signing algorithm '%s' but the requested OAuth 2.0 Client enforces signing algorithm '%s'.", token.Header[JWTHeaderKeyAlgorithm], fclient.GetTokenEndpointAuthSigningAlgorithm()))
		}

		switch token.Method {
		case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
			return p.findClientPublicJWK(ctx, fclient, token, true)
		case jwt.SigningMethodES256, jwt.SigningMethodES384, jwt.SigningMethodES512:
			return p.findClientPublicJWK(ctx, fclient, token, false)
		case jwt.SigningMethodPS256, jwt.SigningMethodPS384, jwt.SigningMethodPS512:
			return p.findClientPublicJWK(ctx, fclient, token, true)
		case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
			var secret *plaintext.Digest

			if secret, ok = fclient.Secret.PlainText(); ok {
				return secret.Key(), nil
			}

			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("This client does not support authentication method 'client_secret_jwt' as the client secret is not in plaintext."))
		default:
			return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHintf("The 'client_assertion' request parameter uses unsupported signing algorithm '%s'.", token.Header[JWTHeaderKeyAlgorithm]))
		}
	}, parserOpts...)

	if err != nil {
		return
	}

	return client, token, token.Claims.(jwt.MapClaims), err
}

func parseJWTAssertionClientID(clientID string, claims jwt.MapClaims) (string, error) {
	if len(clientID) > 0 {
		return clientID, nil
	}

	var err error

	if clientID, err = claims.GetSubject(); err != nil {
		return "", errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'sub' from 'client_assertion' is invalid.").WithWrap(err))
	} else if len(clientID) > 0 {
		return clientID, nil
	}

	if clientID, err = claims.GetIssuer(); err != nil {
		return "", errorsx.WithStack(fosite.ErrInvalidClient.WithHint("Claim 'iss' from 'client_assertion' is invalid.").WithWrap(err))
	} else if len(clientID) > 0 {
		return clientID, nil
	}

	return "", fosite.ErrInvalidClient.WithHint("There was insufficient information in the request to identify the client this request is for.")
}

func (p *OpenIDConnectProvider) checkClientSecret(ctx context.Context, client Client, value []byte) (err error) {
	if len(value) == 0 {
		if client.IsPublic() {
			return errorsx.WithStack(fosite.ErrInvalidClient.WithHint("The registered client doesn't support the authentication method used."))
		}

		return errorsx.WithStack(fosite.ErrInvalidRequest.WithHint("The registered client requires authentication but it was not found."))
	}

	secret := client.GetSecret()

	if secret == nil {
		return errorsx.WithStack(fosite.ErrInvalidClient.WithHint("The registered client doesn't support the authentication method used."))
	}

	if secret.MatchBytes(value) {
		return nil
	}

	err = errPasswordsDoNotMatch

	cc, ok := client.(fosite.ClientWithSecretRotation)
	if !ok {
		return err
	}

	for _, hash := range cc.GetRotatedHashes() {
		if err = p.Config.GetSecretsHasher(ctx).Compare(ctx, hash, value); err == nil {
			return nil
		}
	}

	return err
}

func (p *OpenIDConnectProvider) findClientPublicJWK(ctx context.Context, oidcClient fosite.OpenIDConnectClient, t *jwt.Token, expectsRSAKey bool) (any, error) {
	if set := oidcClient.GetJSONWebKeys(); set != nil {
		return findPublicKey(t, set, expectsRSAKey)
	}

	if location := oidcClient.GetJSONWebKeysURI(); len(location) > 0 {
		keys, err := p.Config.GetJWKSFetcherStrategy(ctx).Resolve(ctx, location, false)
		if err != nil {
			return nil, err
		}

		if key, err := findPublicKey(t, keys, expectsRSAKey); err == nil {
			return key, nil
		}

		keys, err = p.Config.GetJWKSFetcherStrategy(ctx).Resolve(ctx, location, true)
		if err != nil {
			return nil, err
		}

		return findPublicKey(t, keys, expectsRSAKey)
	}

	return nil, errorsx.WithStack(fosite.ErrInvalidClient.WithHint("The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request."))
}

func findPublicKey(t *jwt.Token, set *jose.JSONWebKeySet, expectsRSAKey bool) (any, error) {
	if len(set.Keys) == 0 {
		return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("The retrieved JSON Web Key Set does not contain any key."))
	}

	keys := set.Keys

	kid, ok := t.Header[JWTHeaderKeyIdentifier].(string)
	if ok {
		keys = set.Key(kid)
	}

	if len(keys) == 0 {
		return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("The JSON Web Token uses signing key with kid '%s', which could not be found.", kid))
	}

	for _, key := range keys {
		if key.Use != KeyUseSignature {
			continue
		}

		if expectsRSAKey {
			if k, ok := key.Key.(*rsa.PublicKey); ok {
				return k, nil
			}
		} else {
			if k, ok := key.Key.(*ecdsa.PublicKey); ok {
				return k, nil
			}
		}
	}

	if expectsRSAKey {
		return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("Unable to find RSA public key with a 'use' value of 'sig' for kid '%s' in JSON Web Key Set.", kid))
	} else {
		return nil, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("Unable to find ECDSA public key with a 'use' value of 'sig' for kid '%s' in JSON Web Key Set.", kid))
	}
}

func clientCredentialsFromRequest(header http.Header, form url.Values) (clientID, clientSecret, method string, err error) {
	var ok bool

	switch clientID, clientSecret, ok, err = clientCredentialsFromBasicAuth(header); {
	case err != nil:
		return "", "", "", errorsx.WithStack(fosite.ErrInvalidRequest.WithHint("The client credentials in the HTTP authorization header could not be parsed. Either the scheme was missing, the scheme was invalid, or the value had malformed data.").WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	case ok:
		return clientID, clientSecret, ClientAuthMethodClientSecretBasic, nil
	default:
		clientID, clientSecret = form.Get(FormParameterClientID), form.Get(FormParameterClientSecret)

		switch {
		case clientID == "":
			return "", "", "", errorsx.WithStack(fosite.ErrInvalidRequest.WithHint("Client credentials missing or malformed in both HTTP Authorization header and HTTP POST body."))
		case clientSecret == "":
			return clientID, "", ClientAuthMethodNone, nil
		default:
			return clientID, clientSecret, ClientAuthMethodClientSecretPost, nil
		}
	}
}

func clientCredentialsFromBasicAuth(header http.Header) (clientID, clientSecret string, ok bool, err error) {
	auth := header.Get(fasthttp.HeaderAuthorization)

	if auth == "" {
		return "", "", false, nil
	}

	scheme, value, ok := strings.Cut(auth, " ")

	if !ok {
		return "", "", false, errors.New("failed to parse http authorization header: invalid scheme: the scheme was missing")
	}

	if !strings.EqualFold(scheme, httpAuthSchemeBasic) {
		return "", "", false, fmt.Errorf("failed to parse http authorization header: invalid scheme: expected the %s scheme but received %s", httpAuthSchemeBasic, scheme)
	}

	c, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to parse http authorization header: invalid value: malformed base64 data: %w", err)
	}

	cs := string(c)

	clientID, clientSecret, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false, errors.New("failed to parse http authorization header: invalid value: the basic scheme separator was missing")
	}

	return clientID, clientSecret, true, nil
}

// PKCEHandler is a fork of pkce.Handler with modifications to rectify bugs. It implements the
// fosite.TokenEndpointHandler.
type PKCEHandler struct {
	AuthorizeCodeStrategy oauth2.AuthorizeCodeStrategy
	Storage               pkce.PKCERequestStorage
	Config                interface {
		fosite.EnforcePKCEProvider
		fosite.EnforcePKCEForPublicClientsProvider
		fosite.EnablePKCEPlainChallengeMethodProvider
	}
}

var (
	rePKCEVerifier = regexp.MustCompile(`[^\w.\-~]`)
)

// HandleAuthorizeEndpointRequest implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) HandleAuthorizeEndpointRequest(ctx context.Context, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder) error {
	// This let's us define multiple response types, for example open id connect's id_token.
	if !requester.GetResponseTypes().Has(FormParameterAuthorizationCode) {
		return nil
	}

	challenge := requester.GetRequestForm().Get(FormParameterCodeChallenge)
	method := requester.GetRequestForm().Get(FormParameterCodeChallengeMethod)
	client := requester.GetClient()

	if err := c.validate(ctx, challenge, method, client); err != nil {
		return err
	}

	// We don't need a session if it's not enforced and the PKCE parameters are not provided by the client.
	if challenge == "" && method == "" {
		return nil
	}

	code := responder.GetCode()
	if len(code) == 0 {
		return errorsx.WithStack(fosite.ErrServerError.WithDebug("The PKCE handler must be loaded after the authorize code handler."))
	}

	signature := c.AuthorizeCodeStrategy.AuthorizeCodeSignature(ctx, code)

	if err := c.Storage.CreatePKCERequestSession(ctx, signature, requester.Sanitize([]string{FormParameterCodeChallenge, FormParameterCodeChallengeMethod})); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	return nil
}

func (c *PKCEHandler) validate(ctx context.Context, challenge, method string, client fosite.Client) (err error) {
	if len(challenge) == 0 {
		// If the server requires Proof Key for Code Exchange (PKCE) by OAuth
		// clients and the client does not send the "code_challenge" in
		// the request, the authorization endpoint MUST return the authorization
		// error response with the "error" value set to "invalid_request".  The
		// "error_description" or the response of "error_uri" SHOULD explain the
		// nature of error, e.g., code challenge required.
		return c.validateNoPKCE(ctx, client)
	}

	// If the server supporting PKCE does not support the requested
	// transformation, the authorization endpoint MUST return the
	// authorization error response with "error" value set to
	// "invalid_request".  The "error_description" or the response of
	// "error_uri" SHOULD explain the nature of error, e.g., transform
	// algorithm not supported.
	switch method {
	case PKCEChallengeMethodSHA256:
		break
	case PKCEChallengeMethodPlain:
		fallthrough
	case "":
		if !c.Config.GetEnablePKCEPlainChallengeMethod(ctx) {
			return errorsx.WithStack(fosite.ErrInvalidRequest.
				WithHint("Clients must use the 'S256' PKCE 'code_challenge_method' but the 'plain' method was requested.").
				WithDebug("The server is configured in a way that enforces PKCE 'S256' as challenge method for clients."))
		}
	default:
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithHint("The code_challenge_method is not supported, use S256 instead."))
	}

	return nil
}

func (c *PKCEHandler) validateNoPKCE(ctx context.Context, client fosite.Client) error {
	var enforce bool

	enforce = c.Config.GetEnforcePKCE(ctx)

	if enforce {
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithHint("Clients must include a code_challenge when performing the authorize code flow, but it is missing.").
			WithDebug("The server is configured in a way that enforces PKCE for clients."))
	}

	enforce = c.Config.GetEnforcePKCEForPublicClients(ctx)
	public := client.IsPublic()

	if enforce && public {
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithHint("This client must include a code_challenge when performing the authorize code flow, but it is missing.").
			WithDebug("The server is configured in a way that enforces PKCE for this client."))
	}

	return nil
}

// HandleTokenEndpointRequest implements fosite.TokenEndpointHandler partially.
//
//nolint:gocyclo
func (c *PKCEHandler) HandleTokenEndpointRequest(ctx context.Context, requester fosite.AccessRequester) (err error) {
	if !c.CanHandleTokenEndpointRequest(ctx, requester) {
		return errorsx.WithStack(fosite.ErrUnknownRequest)
	}

	// code_verifier
	// REQUIRED.  Code verifier
	//
	// The "code_challenge_method" is bound to the Authorization Code when
	// the Authorization Code is issued.  That is the method that the token
	// endpoint MUST use to verify the "code_verifier".
	verifier := requester.GetRequestForm().Get(FormParameterCodeVerifier)

	code := requester.GetRequestForm().Get(FormParameterAuthorizationCode)
	signature := c.AuthorizeCodeStrategy.AuthorizeCodeSignature(ctx, code)
	pkceRequest, err := c.Storage.GetPKCERequestSession(ctx, signature, requester.GetSession())

	nv := len(verifier)

	if errors.Is(err, fosite.ErrNotFound) {
		if nv == 0 {
			return c.validateNoPKCE(ctx, requester.GetClient())
		}

		return errorsx.WithStack(fosite.ErrInvalidGrant.WithHint("Unable to find initial PKCE data tied to this request.").WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	} else if err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	if err = c.Storage.DeletePKCERequestSession(ctx, signature); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	challenge := pkceRequest.GetRequestForm().Get(FormParameterCodeChallenge)
	method := pkceRequest.GetRequestForm().Get(FormParameterCodeChallengeMethod)
	client := pkceRequest.GetClient()

	if err = c.validate(ctx, challenge, method, client); err != nil {
		return err
	}

	nc := len(challenge)

	if !c.Config.GetEnforcePKCE(ctx) && nc == 0 && nv == 0 {
		return nil
	}

	// NOTE: The code verifier SHOULD have enough entropy to make it
	// 	impractical to guess the value.  It is RECOMMENDED that the output of
	// 	a suitable random number generator be used to create a 32-octet
	// 	sequence.  The octet sequence is then base64url-encoded to produce a
	// 	43-octet URL safe string to use as the code verifier.

	// Miscellaneous validations.
	switch {
	case nv < 43:
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier must be at least 43 characters."))
	case nv > 128:
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier can not be longer than 128 characters."))
	case rePKCEVerifier.MatchString(verifier):
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier must only contain [a-Z], [0-9], '-', '.', '_', '~'."))
	case nc == 0:
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier was provided but the code challenge was absent from the authorization request."))
	}

	// Upon receipt of the request at the token endpoint, the server
	// verifies it by calculating the code challenge from the received
	// "code_verifier" and comparing it with the previously associated
	// "code_challenge", after first transforming it according to the
	// "code_challenge_method" method specified by the client.
	//
	// 	If the "code_challenge_method" from Section 4.3 was "S256", the
	// received "code_verifier" is hashed by SHA-256, base64url-encoded, and
	// then compared to the "code_challenge", i.e.:
	//
	// BASE64URL-ENCODE(SHA256(ASCII(code_verifier))) == code_challenge
	//
	// If the "code_challenge_method" from Section 4.3 was "plain", they are
	// compared directly, i.e.:
	//
	// code_verifier == code_challenge.
	//
	// 	If the values are equal, the token endpoint MUST continue processing
	// as normal (as defined by OAuth 2.0 [RFC6749]).  If the values are not
	// equal, an error response indicating "invalid_grant" as described in
	// Section 5.2 of [RFC6749] MUST be returned.
	switch method {
	case PKCEChallengeMethodSHA256:
		hash := sha256.New()
		if _, err = hash.Write([]byte(verifier)); err != nil {
			return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
		}

		sum := hash.Sum([]byte{})

		expected := make([]byte, base64.RawURLEncoding.EncodedLen(len(sum)))

		base64.RawURLEncoding.Strict().Encode(expected, sum)

		if subtle.ConstantTimeCompare(expected, []byte(challenge)) == 0 {
			return errorsx.WithStack(fosite.ErrInvalidGrant.
				WithHint("The PKCE code challenge did not match the code verifier."))
		}
	case PKCEChallengeMethodPlain:
		fallthrough
	default:
		if subtle.ConstantTimeCompare([]byte(verifier), []byte(challenge)) == 0 {
			return errorsx.WithStack(fosite.ErrInvalidGrant.
				WithHint("The PKCE code challenge did not match the code verifier."))
		}
	}

	return nil
}

// PopulateTokenEndpointResponse implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) PopulateTokenEndpointResponse(ctx context.Context, requester fosite.AccessRequester, responder fosite.AccessResponder) (err error) {
	return nil
}

// CanSkipClientAuth implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) CanSkipClientAuth(ctx context.Context, requester fosite.AccessRequester) bool {
	return false
}

// CanHandleTokenEndpointRequest implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) CanHandleTokenEndpointRequest(ctx context.Context, requester fosite.AccessRequester) bool {
	return requester.GetGrantTypes().ExactOne(GrantTypeAuthorizationCode)
}

var (
	_ fosite.TokenEndpointHandler = (*PKCEHandler)(nil)
)
