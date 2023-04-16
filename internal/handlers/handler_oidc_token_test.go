package handlers

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestClientAuthenticationStrategySuite(t *testing.T) {
	suite.Run(t, &ClientAuthenticationStrategySuite{})
}

type ClientAuthenticationStrategySuite struct {
	suite.Suite

	issuerURL *url.URL

	ctrl     *gomock.Controller
	store    *mocks.MockStorage
	provider *oidc.OpenIDConnectProvider
}

func (s *ClientAuthenticationStrategySuite) GetIssuerURL() *url.URL {
	if s.issuerURL == nil {
		s.issuerURL = MustParseRequestURI("https://auth.example.com")
	}

	return s.issuerURL
}

func (s *ClientAuthenticationStrategySuite) GetTokenURL() *url.URL {
	return s.GetIssuerURL().JoinPath(oidc.EndpointPathToken)
}

func (s *ClientAuthenticationStrategySuite) GetRequest(values *url.Values) (r *http.Request) {
	var err error

	r, err = http.NewRequest(http.MethodPost, s.GetTokenURL().String(), strings.NewReader(values.Encode()))

	s.Require().NoError(err)
	s.Require().NotNil(r)

	r.Header.Set(fasthttp.HeaderContentType, "application/x-www-form-urlencoded")

	s.Require().NoError(r.ParseForm())

	return r
}

func (s *ClientAuthenticationStrategySuite) GetAssertionValues(token string) *url.Values {
	values := &url.Values{}

	values.Set(oidc.FormParameterClientAssertionType, oidc.ClientAssertionJWTBearerType)

	if token != "" {
		values.Set(oidc.FormParameterClientAssertion, token)
	}

	return values
}

func (s *ClientAuthenticationStrategySuite) GetAssertionRequest(token string) (r *http.Request) {
	values := s.GetAssertionValues(token)

	return s.GetRequest(values)
}

func (s *ClientAuthenticationStrategySuite) GetCtx() oidc.OpenIDConnectContext {
	return &oidc.MockOpenIDConnectContext{
		Context:       context.Background(),
		MockIssuerURL: s.GetIssuerURL(),
	}
}

func (s *ClientAuthenticationStrategySuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mocks.NewMockStorage(s.ctrl)

	var err error

	secret := MustDecodeSecret("$plaintext$client-secret")

	s.provider, err = oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       MustParseRSAPrivateKey(exampleIssuerPrivateKey),
		HMACSecret:             "abc123",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "hs512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmHMACWithSHA512,
			},
			{
				ID:     "hs5122",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmHMACWithSHA512,
			},
			{
				ID:     "hs256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmHMACWithSHA256,
			},
			{
				ID:     "rs256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmRSAWithSHA256,
			},
			{
				ID:     "hashed",
				Secret: MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng"),
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmHMACWithSHA512,
			},
			{
				ID:     oidc.ClientAuthMethodClientSecretBasic,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretBasic,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmNone,
			},
			{
				ID:     oidc.ClientAuthMethodNone,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodNone,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmNone,
			},
			{
				ID:     oidc.ClientAuthMethodClientSecretPost,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretPost,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmNone,
			},
			{
				ID:     "bad_method",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:           "bad_method",
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmNone,
			},
			{
				ID:     "base",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
		},
	}, s.store, nil)

	s.Require().NoError(err)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateJWT() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	sig := fmt.Sprintf("%x", sha256.Sum256([]byte(assertion.ID)))

	ctx := s.GetCtx()

	gomock.InOrder(
		s.store.
			EXPECT().LoadOAuth2BlacklistedJTI(ctx, sig).
			Return(nil, sql.ErrNoRows),

		s.store.
			EXPECT().SaveOAuth2BlacklistedJTI(ctx, model.OAuth2BlacklistedJTI{Signature: sig, ExpiresAt: assertion.ExpiresAt.Time}).
			Return(nil),
	)

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.NoError(ErrorToRFC6749ErrorTest(err))
	s.Require().NotNil(client)
	s.Equal("hs512", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnJTIKnown() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	sig := fmt.Sprintf("%x", sha256.Sum256([]byte(assertion.ID)))

	ctx := s.GetCtx()

	gomock.InOrder(
		s.store.
			EXPECT().LoadOAuth2BlacklistedJTI(ctx, sig).
			Return(nil, sql.ErrNoRows),

		s.store.
			EXPECT().SaveOAuth2BlacklistedJTI(ctx, model.OAuth2BlacklistedJTI{Signature: sig, ExpiresAt: assertion.ExpiresAt.Time}).
			Return(fosite.ErrJTIKnown),
	)

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "The jti was already used.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateJWTWithArbitraryClaims() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	a["aaa"] = "abc"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	sig := fmt.Sprintf("%x", sha256.Sum256([]byte(assertion.ID)))

	ctx := s.GetCtx()

	gomock.InOrder(
		s.store.
			EXPECT().LoadOAuth2BlacklistedJTI(ctx, sig).
			Return(nil, sql.ErrNoRows),

		s.store.
			EXPECT().SaveOAuth2BlacklistedJTI(ctx, model.OAuth2BlacklistedJTI{Signature: sig, ExpiresAt: assertion.ExpiresAt.Time}).
			Return(nil),
	)

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.NoError(ErrorToRFC6749ErrorTest(err))
	s.Require().NotNil(client)
	s.Equal("hs512", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingSubClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	delete(a, oidc.ClaimSubject)

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The claim 'sub' from the client_assertion JSON Web Token is undefined.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithInvalidExpClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	a[oidc.ClaimExpirationTime] = "not a number"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token is expired.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingIssClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	delete(a, oidc.ClaimIssuer)

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'iss' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithInvalidAudClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertion.Audience = []string{"notvalid"}

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	sig := fmt.Sprintf("%x", sha256.Sum256([]byte(assertion.ID)))

	ctx := s.GetCtx()

	gomock.InOrder(
		s.store.
			EXPECT().LoadOAuth2BlacklistedJTI(ctx, sig).
			Return(nil, sql.ErrNoRows),

		s.store.
			EXPECT().SaveOAuth2BlacklistedJTI(ctx, model.OAuth2BlacklistedJTI{Signature: sig, ExpiresAt: assertion.ExpiresAt.Time}).
			Return(nil),
	)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'audience' from 'client_assertion' must match the authorization server's token endpoint 'https://auth.example.com/api/oidc/token'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithInvalidAssertionType() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	v := s.GetAssertionValues(token)

	v.Set(oidc.FormParameterClientAssertionType, "not_valid")

	r := s.GetRequest(v)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Unknown client_assertion_type 'not_valid'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingJTIClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	delete(a, oidc.ClaimJWTID)

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'jti' from 'client_assertion' must be set but is not.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMismatchedIssClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertion.Issuer = "hs256"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'iss' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMismatchedFormClientID() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertion.Issuer = "hs5122"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	values := s.GetAssertionValues(token)

	values.Set(oidc.FormParameterClientID, "hs5122")

	r := s.GetRequest(values)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'sub' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMismatchedFormClientIDWithIss() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	values := s.GetAssertionValues(token)

	values.Set(oidc.FormParameterClientID, "hs5122")

	r := s.GetRequest(values)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'iss' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingClient() {
	assertion := NewAssertion("noclient", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). invalid_client")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailBadSecret() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret-wrong"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The signature is invalid.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailMethodNone() {
	assertion := NewAssertion(oidc.ClientAuthMethodNone, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client does not support client authentication, however 'client_assertion' was provided in the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionMethodClientSecretPost() {
	assertion := NewAssertion(oidc.ClientAuthMethodClientSecretPost, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client only supports client authentication method 'client_secret_post', however 'client_assertion' was provided in the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionMethodBad() {
	assertion := NewAssertion("bad_method", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client only supports client authentication method 'bad_method', however that method is not supported by this server.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionBaseClient() {
	assertion := NewAssertion("base", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The client configuration does not support OpenID Connect specific authentication methods.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionMethodClientSecretBasic() {
	assertion := NewAssertion(oidc.ClientAuthMethodClientSecretBasic, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client only supports client authentication method 'client_secret_basic', however 'client_assertion' was provided in the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailHashedSecret() {
	assertion := NewAssertion("hashed", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This client does not support authentication method 'client_secret_jwt' as the client secret is not in plaintext.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailExpiredToken() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Minute*-3), time.Unix(time.Now().Add(time.Minute*-1).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token is expired.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailNotYetValid() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Minute*-3), time.Unix(time.Now().Add(time.Minute*1).Unix(), 0))

	assertion.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Second * 10))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	s.NoError(r.ParseForm())

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token isn't valid yet.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailTokenUsedBeforeIssued() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Minute*3), time.Unix(time.Now().Add(time.Minute*8).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token was used before it was issued.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailMalformed() {
	r := s.GetAssertionRequest("bad token")

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token is malformed.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailMissingAssertion() {
	r := s.GetAssertionRequest("")

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The client_assertion request parameter must be set when using client_assertion_type of 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer'.")
	s.Nil(client)
}

/*

func TestOpenIDConnectProvider_DefaultClientAuthenticationStrategy_ShouldValidateJWT(t *testing.T) {
	aClientID := "a-client"
	aClientSecret := "a-client-secret"
	root := MustParseRequestURI("https://auth.example.com")
	tokenURL := root.JoinPath(oidc.EndpointPathToken)

	require.NotNil(t, tokenURL)

	ctrl := gomock.NewController(t)

	store := mocks.NewMockStorage(ctrl)

	provider, err := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       MustParseRSAPrivateKey(exampleIssuerPrivateKey),
		HMACSecret:             "abc123",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     aClientID,
				Secret: MustDecodeSecret("$plaintext$" + aClientSecret),
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://google.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmHMACWithSHA512,
			},
		},
	}, store, nil)

	assert.NotNil(t, provider)
	assert.NoError(t, err)

	assertion := NewAssertion(aClientID, tokenURL, time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jose.HS512, assertion.ToMapClaims())

	token, err := assertionJWT.SignedString([]byte(aClientSecret))

	require.NoError(t, err)
	require.NotEqual(t, "", token)

	values := &url.Values{}

	values.Set(oidc.FormParameterClientAssertionType, oidc.ClientAssertionJWTBearerType)
	values.Set(oidc.FormParameterClientAssertion, token)

	var ctx oidc.OpenIDConnectContext = &oidc.MockOpenIDConnectContext{
		Context:       context.Background(),
		MockIssuerURL: MustParseRequestURI("https://auth.example.com"),
	}

	r, err := http.NewRequest(http.MethodPost, tokenURL.String(), strings.NewReader(values.Encode()))

	assert.NoError(t, err)

	require.NotNil(t, r)

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	assert.NoError(t, r.ParseForm())

	sig := fmt.Sprintf("%x", sha256.Sum256([]byte(assertion.JTI)))

	gomock.InOrder(
		store.
			EXPECT().LoadOAuth2BlacklistedJTI(ctx, sig).
			Return(nil, sql.ErrNoRows),

		store.
			EXPECT().SaveOAuth2BlacklistedJTI(ctx, model.OAuth2BlacklistedJTI{Signature: sig, ExpiresAt: assertion.ExpiresAt}).
			Return(nil),
	)

	client, err := provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	assert.NoError(t, ErrorToRFC6749ErrorTest(err))

	require.NotNil(t, client)

	assert.Equal(t, aClientID, client.GetID())
}

func TestOpenIDConnectProvider_DefaultClientAuthenticationStrategy_ShouldErrorInvalidKTY(t *testing.T) {
	aClientID := "a-client"
	aClientSecret := "a-client-secret"
	root := MustParseRequestURI("https://auth.example.com")
	tokenURL := root.JoinPath(oidc.EndpointPathToken)

	require.NotNil(t, tokenURL)

	provider, err := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       MustParseRSAPrivateKey(exampleIssuerPrivateKey),
		HMACSecret:             "abc123",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     aClientID,
				Secret: MustDecodeSecret("$plaintext$" + aClientSecret),
				Policy: "one_factor",
				RedirectURIs: []string{
					"https://google.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmHMACWithSHA256,
			},
		},
	}, nil, nil)

	assert.NotNil(t, provider)
	assert.NoError(t, err)

	assertion := NewAssertion(aClientID, tokenURL, time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jose.HS512, assertion.ToMapClaims())

	token, err := assertionJWT.SignedString([]byte(aClientSecret))

	require.NoError(t, err)
	require.NotEqual(t, "", token)

	values := &url.Values{}

	values.Set(oidc.FormParameterClientAssertionType, oidc.ClientAssertionJWTBearerType)
	values.Set(oidc.FormParameterClientAssertion, token)

	var ctx oidc.OpenIDConnectContext = &oidc.MockOpenIDConnectContext{
		Context:       context.Background(),
		MockIssuerURL: MustParseRequestURI("https://auth.example.com"),
	}

	r, err := http.NewRequest(http.MethodPost, tokenURL.String(), strings.NewReader(values.Encode()))

	assert.NoError(t, err)

	require.NotNil(t, r)

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	assert.NoError(t, r.ParseForm())

	client, err := provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	assert.Nil(t, client)
	assert.EqualError(t, err, "invalid_client")
	assert.EqualError(t, ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'HS512' but the requested OAuth 2.0 Client enforces signing algorithm 'HS256'.")
}

func TestOpenIDConnectProvider_DefaultClientAuthenticationStrategy_ShouldErrorEmptyAssertion(t *testing.T) {
	aClientID := "a-client"
	aClientSecret := "a-client-secret"
	root := MustParseRequestURI("https://auth.example.com")
	tokenURL := root.JoinPath(oidc.EndpointPathToken)

	require.NotNil(t, tokenURL)

	provider, err := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       MustParseRSAPrivateKey(exampleIssuerPrivateKey),
		HMACSecret:             "abc123",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     aClientID,
				Secret: MustDecodeSecret("$plaintext$" + aClientSecret),
				Policy: "one_factor",
				RedirectURIs: []string{
					"https://google.com",
				},
				TokenEndpointAuthMethod:           oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlgorithm: oidc.SigningAlgorithmHMACWithSHA256,
			},
		},
	}, nil, nil)

	assert.NotNil(t, provider)
	assert.NoError(t, err)

	values := &url.Values{}

	values.Set(oidc.FormParameterClientAssertionType, oidc.ClientAssertionJWTBearerType)

	var ctx oidc.OpenIDConnectContext = &oidc.MockOpenIDConnectContext{
		Context:       context.Background(),
		MockIssuerURL: MustParseRequestURI("https://auth.example.com"),
	}

	r, err := http.NewRequest(http.MethodPost, tokenURL.String(), strings.NewReader(values.Encode()))

	assert.NoError(t, err)

	require.NotNil(t, r)

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	assert.NoError(t, r.ParseForm())

	client, err := provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	assert.Nil(t, client)
	assert.EqualError(t, err, "invalid_request")
	assert.EqualError(t, ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The client_assertion request parameter must be set when using client_assertion_type of 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer'.")
}

temp comment.
*/

type RegisteredClaims struct {
	jwt.RegisteredClaims
}

func (r *RegisteredClaims) ToMapClaims() jwt.MapClaims {
	claims := jwt.MapClaims{}

	if r.ID != "" {
		claims[oidc.ClaimJWTID] = r.ID
	}

	if r.Subject != "" {
		claims[oidc.ClaimSubject] = r.Subject
	}

	if r.Issuer != "" {
		claims[oidc.ClaimIssuer] = r.Issuer
	}

	if len(r.Audience) != 0 {
		claims[oidc.ClaimAudience] = r.Audience
	}

	if r.NotBefore != nil {
		claims[oidc.ClaimNotBefore] = r.NotBefore
	}

	if r.ExpiresAt != nil {
		claims[oidc.ClaimExpirationTime] = r.ExpiresAt
	}

	if r.IssuedAt != nil {
		claims[oidc.ClaimIssuedAt] = r.IssuedAt
	}

	return claims
}

func NewAssertion(clientID string, tokenURL *url.URL, iat, exp time.Time) RegisteredClaims {
	return RegisteredClaims{
		jwt.RegisteredClaims{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Issuer: clientID,
			Audience: []string{
				tokenURL.String(),
			},
			Subject:   clientID,
			IssuedAt:  jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
}

type RFC6749ErrorTest struct {
	*fosite.RFC6749Error
}

func (err *RFC6749ErrorTest) Error() string {
	return err.WithExposeDebug(true).GetDescription()
}

func ErrorToRFC6749ErrorTest(err error) (rfc error) {
	if err == nil {
		return nil
	}

	ferr := fosite.ErrorToRFC6749Error(err)

	return &RFC6749ErrorTest{ferr}
}

func MustDecodeSecret(value string) *schema.PasswordDigest {
	if secret, err := schema.DecodePasswordDigest(value); err != nil {
		panic(err)
	} else {
		return secret
	}
}

func MustParseRequestURI(input string) *url.URL {
	if requestURI, err := url.ParseRequestURI(input); err != nil {
		panic(err)
	} else {
		return requestURI
	}
}

func MustParseRSAPrivateKey(data string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "RSA PRIVATE KEY" {
		panic("not private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

const exampleIssuerPrivateKey = "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAvcMVMB2vEbqI6PlSNJ4HmUyMxBDJ5iY7FS+zDDAHOZBg9S3S\nKcAn1CZcnyL0VvJ7wcdhR6oTnOwR94eKvzUyJZ+GL2hTMm27dubEYsNdhoCl6N3X\nyEEohNfoxiiCYraVauX8X3M9jFzbEz9+pacaDbHB2syaJ1qFmMNR+HSu2jPzOo7M\nlqKIOgUzA0741MaYNt47AEVg4XU5ORLdolbAkItmYg1QbyFndg9H5IvwKkYaXTGE\nlgDBcPUC0yVjAC15Mguquq+jZeQay+6PSbHTD8PQMOkLjyChI2xEhVNbdCXe676R\ncMW2R/gjrcK23zmtmTWRfdC1iZLSlHO+bJj9vQIDAQABAoIBAEZvkP/JJOCJwqPn\nV3IcbmmilmV4bdi1vByDFgyiDyx4wOSA24+PubjvfFW9XcCgRPuKjDtTj/AhWBHv\nB7stfa2lZuNV7/u562mZArA+IAr62Zp0LdIxDV8x3T8gbjVB3HhPYbv0RJZDKTYd\nzV6jhfIrVu9mHpoY6ZnodhapCPYIyk/d49KBIHZuAc25CUjMXgTeaVtf0c996036\nUxW6ef33wAOJAvW0RCvbXAJfmBeEq2qQlkjTIlpYx71fhZWexHifi8Ouv3Zonc+1\n/P2Adq5uzYVBT92f9RKHg9QxxNzVrLjSMaxyvUtWQCAQfW0tFIRdqBGsHYsQrFtI\nF4yzv8ECgYEA7ntpyN9HD9Z9lYQzPCR73sFCLM+ID99aVij0wHuxK97bkSyyvkLd\n7MyTaym3lg1UEqWNWBCLvFULZx7F0Ah6qCzD4ymm3Bj/ADpWWPgljBI0AFml+HHs\nhcATmXUrj5QbLyhiP2gmJjajp1o/rgATx6ED66seSynD6JOH8wUhhZUCgYEAy7OA\n06PF8GfseNsTqlDjNF0K7lOqd21S0prdwrsJLiVzUlfMM25MLE0XLDUutCnRheeh\nIlcuDoBsVTxz6rkvFGD74N+pgXlN4CicsBq5ofK060PbqCQhSII3fmHobrZ9Cr75\nHmBjAxHx998SKaAAGbBbcYGUAp521i1pH5CEPYkCgYEAkUd1Zf0+2RMdZhwm6hh/\nrW+l1I6IoMK70YkZsLipccRNld7Y9LbfYwYtODcts6di9AkOVfueZJiaXbONZfIE\nZrb+jkAteh9wGL9xIrnohbABJcV3Kiaco84jInUSmGDtPokncOENfHIEuEpuSJ2b\nbx1TuhmAVuGWivR0+ULC7RECgYEAgS0cDRpWc9Xzh9Cl7+PLsXEvdWNpPsL9OsEq\n0Ep7z9+/+f/jZtoTRCS/BTHUpDvAuwHglT5j3p5iFMt5VuiIiovWLwynGYwrbnNS\nqfrIrYKUaH1n1oDS+oBZYLQGCe9/7EifAjxtjYzbvSyg//SPG7tSwfBCREbpZXj2\nqSWkNsECgYA/mCDzCTlrrWPuiepo6kTmN+4TnFA+hJI6NccDVQ+jvbqEdoJ4SW4L\nzqfZSZRFJMNpSgIqkQNRPJqMP0jQ5KRtJrjMWBnYxktwKz9fDg2R2MxdFgMF2LH2\nHEMMhFHlv8NDjVOXh1KwRoltNGVWYsSrD9wKU9GhRCEfmNCGrvBcEg==\n-----END RSA PRIVATE KEY-----"
