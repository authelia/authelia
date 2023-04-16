package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
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

func (s *ClientAuthenticationStrategySuite) GetBaseRequest(body io.Reader) (r *http.Request) {
	var err error

	r, err = http.NewRequest(http.MethodPost, s.GetTokenURL().String(), body)

	s.Require().NoError(err)
	s.Require().NotNil(r)

	r.Header.Set(fasthttp.HeaderContentType, "application/x-www-form-urlencoded")

	return r
}

func (s *ClientAuthenticationStrategySuite) GetRequest(values *url.Values) (r *http.Request) {
	var body io.Reader

	if values != nil {
		body = strings.NewReader(values.Encode())
	}

	r = s.GetBaseRequest(body)

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

func (s *ClientAuthenticationStrategySuite) GetClientValues(id string) *url.Values {
	values := &url.Values{}

	values.Set(oidc.FormParameterClientID, id)

	return values
}

func (s *ClientAuthenticationStrategySuite) GetClientValuesPost(id, secret string) *url.Values {
	values := s.GetClientValues(id)

	values.Set(oidc.FormParameterClientSecret, secret)

	return values
}

func (s *ClientAuthenticationStrategySuite) GetClientSecretBasicRequest(id, secret string) (r *http.Request) {
	values := s.GetClientValues(id)

	r = s.GetRequest(values)

	r.SetBasicAuth(id, secret)

	return r
}

func (s *ClientAuthenticationStrategySuite) GetClientSecretPostRequest(id, secret string) (r *http.Request) {
	values := s.GetClientValuesPost(id, secret)

	return s.GetRequest(values)
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
		IssuerPrivateKey:       MustParseRSAPrivateKey(exampleRSAPrivateKey),
		HMACSecret:             "abc123",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "hs256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA256,
			},
			{
				ID:     "hs384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA384,
			},
			{
				ID:     "hs512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA512,
			},
			{
				ID:     "rs256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgRSAUsingSHA256,
			},
			{
				ID:     "rs384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgRSAUsingSHA384,
			},
			{
				ID:     "rs512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgRSAUsingSHA512,
			},
			{
				ID:     "ps256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgRSAPSSUsingSHA256,
			},
			{
				ID:     "ps384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgRSAPSSUsingSHA384,
			},
			{
				ID:     "ps512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgRSAPSSUsingSHA512,
			},
			{
				ID:     "es256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgECDSAUsingP256AndSHA256,
			},
			{
				ID:     "es384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgECDSAUsingP384AndSHA384,
			},
			{
				ID:     "es512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgECDSAUsingP521AndSHA512,
			},
			{
				ID:     "hs5122",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA512,
			},
			{
				ID:     "hashed",
				Secret: MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng"),
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA512,
			},
			{
				ID:     oidc.ClientAuthMethodClientSecretBasic,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretBasic,
				TokenEndpointAuthSigningAlg: oidc.SigAlgNone,
			},
			{
				ID:     oidc.ClientAuthMethodNone,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodNone,
				TokenEndpointAuthSigningAlg: oidc.SigAlgNone,
			},
			{
				ID:     oidc.ClientAuthMethodClientSecretPost,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretPost,
				TokenEndpointAuthSigningAlg: oidc.SigAlgNone,
			},
			{
				ID:     "bad_method",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     "bad_method",
				TokenEndpointAuthSigningAlg: oidc.SigAlgNone,
			},
			{
				ID:     "base",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:                      "public",
				Public:                  true,
				Policy:                  authorization.OneFactor.String(),
				TokenEndpointAuthMethod: oidc.ClientAuthMethodNone,
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:     "public-nomethod",
				Public: true,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:                      "public-basic",
				Public:                  true,
				Policy:                  authorization.OneFactor.String(),
				TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretBasic,
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
		},
	}, s.store, nil)

	s.Require().NoError(err)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateAssertionHS256() {
	assertion := NewAssertion("hs256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, assertion)

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
	s.Equal("hs256", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateAssertionHS384() {
	assertion := NewAssertion("hs384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS384, assertion)

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
	s.Equal("hs384", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateAssertionHS512() {
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

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnMismatchedAlg() {
	assertion := NewAssertion("rs256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'HS512' but the requested OAuth 2.0 Client enforces signing algorithm 'RS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnMismatchedAlgSameMethod() {
	assertion := NewAssertion("hs256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'HS512' but the requested OAuth 2.0 Client enforces signing algorithm 'HS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS256() {
	assertion := NewAssertion("rs256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)

	token, err := assertionJWT.SignedString(MustParseRSAPrivateKey(exampleRSAPrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS384() {
	assertion := NewAssertion("rs384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS384, assertion)

	token, err := assertionJWT.SignedString(MustParseRSAPrivateKey(exampleRSAPrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS512() {
	assertion := NewAssertion("rs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS512, assertion)

	token, err := assertionJWT.SignedString(MustParseRSAPrivateKey(exampleRSAPrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS256() {
	assertion := NewAssertion("ps256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)

	token, err := assertionJWT.SignedString(MustParseRSAPrivateKey(exampleRSAPrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS384() {
	assertion := NewAssertion("ps384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS384, assertion)

	token, err := assertionJWT.SignedString(MustParseRSAPrivateKey(exampleRSAPrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS512() {
	assertion := NewAssertion("ps512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS512, assertion)

	token, err := assertionJWT.SignedString(MustParseRSAPrivateKey(exampleRSAPrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES256() {
	assertion := NewAssertion("es256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES256, assertion)

	token, err := assertionJWT.SignedString(MustParseECPrivateKey(exampleECP256PrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES384() {
	assertion := NewAssertion("es384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES384, assertion)

	token, err := assertionJWT.SignedString(MustParseECPrivateKey(exampleECP384PrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES512() {
	assertion := NewAssertion("es512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES512, assertion)

	token, err := assertionJWT.SignedString(MustParseECPrivateKey(exampleECP521PrivateKey))

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
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

func (s *ClientAuthenticationStrategySuite) TestShouldValidateClientSecretPost() {
	r := s.GetClientSecretPostRequest(oidc.ClientAuthMethodClientSecretPost, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.NoError(err)
	s.Require().NotNil(client)
	s.Equal(oidc.ClientAuthMethodClientSecretPost, client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretPostOnClientSecretBasicClient() {
	r := s.GetClientSecretPostRequest(oidc.ClientAuthMethodClientSecretBasic, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client supports client authentication method 'client_secret_basic', but method 'client_secret_post' was requested. You must configure the OAuth 2.0 client's 'token_endpoint_auth_method' value to accept 'client_secret_post'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretPostWrongSecret() {
	r := s.GetClientSecretPostRequest(oidc.ClientAuthMethodClientSecretPost, "client-secret-bad")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The provided client secret did not match the registered client secret.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateClientSecretBasic() {
	r := s.GetClientSecretBasicRequest(oidc.ClientAuthMethodClientSecretBasic, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.NoError(err)
	s.Require().NotNil(client)
	s.Equal(oidc.ClientAuthMethodClientSecretBasic, client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnClientSecretPostWithoutClientID() {
	r := s.GetRequest(&url.Values{oidc.FormParameterClientSecret: []string{"client-secret"}})

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Client credentials missing or malformed in both HTTP Authorization header and HTTP POST body.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnClientSecretBasicWithMalformedClientID() {
	r := s.GetRequest(&url.Values{oidc.FormParameterRequestURI: []string{"not applicable"}})

	r.Header.Set(fasthttp.HeaderAuthorization, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("abc@#%!@#(*%)#@!:client-secret"))))
	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The client id in the HTTP authorization header could not be decoded from 'application/x-www-form-urlencoded'. invalid URL escape '%!@'")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnClientSecretBasicWithMalformedClientSecret() {
	r := s.GetRequest(&url.Values{oidc.FormParameterRequestURI: []string{"not applicable"}})

	r.Header.Set(fasthttp.HeaderAuthorization, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("hs512:abc@#%!@#(*%)#@!"))))
	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The client secret in the HTTP authorization header could not be decoded from 'application/x-www-form-urlencoded'. invalid URL escape '%!@'")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnClientSecretPostClient() {
	r := s.GetClientSecretBasicRequest(oidc.ClientAuthMethodClientSecretPost, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client supports client authentication method 'client_secret_post', but method 'client_secret_basic' was requested. You must configure the OAuth 2.0 client's 'token_endpoint_auth_method' value to accept 'client_secret_basic'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicWrongSecret() {
	r := s.GetClientSecretBasicRequest(oidc.ClientAuthMethodClientSecretBasic, "client-secret-bad")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The provided client secret did not match the registered client secret.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnPublic() {
	r := s.GetClientSecretBasicRequest("public", "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client supports client authentication method 'none', but method 'client_secret_basic' was requested. You must configure the OAuth 2.0 client's 'token_endpoint_auth_method' value to accept 'client_secret_basic'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnPublicWithBasic() {
	r := s.GetClientSecretBasicRequest("public-basic", "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client supports client authentication method 'client_secret_basic', but method 'none' was requested. You must configure the OAuth 2.0 client's 'token_endpoint_auth_method' value to accept 'none'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnInvalidClient() {
	r := s.GetClientSecretBasicRequest("not-a-client", "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). invalid_client")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidatePublic() {
	v := s.GetClientValues("public")

	r := s.GetRequest(v)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.NoError(err)
	s.Require().NotNil(client)
	s.Equal("public", client.GetID())
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
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA512,
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
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA256,
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
				TokenEndpointAuthSigningAlg: oidc.SigAlgHMACUsingSHA256,
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

func MustParseECPrivateKey(data string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "EC PRIVATE KEY" {
		panic("not private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

const exampleRSAPrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA60Vuz1N1wUHiCDIlbz8gE0dWPCmHEWnXKchEEISqIJ6j5Eah
Q/GwX3WK0UV5ATRvWhg6o7/WfrLYcAsi4w79TgMjJHLWIY/jzAS3quEtzOLlLSWZ
9FR9SomQm3T/ETOS8IvSGrksIj0WgX35jB1NnbqSTRnYx7Cg/TBJjmiaqd0b9G/8
LlReaihwGf8tvPgnteWIdon3EI2MKDBkaesRjpL98Cz7VvD7dajseAlUh9jQWVge
sN8qnm8pNPFAYsgxf//Jf0RfsND6H70zKKybDmyct4T4o/8qjivw4ly0XkArDCUj
Qx2KUF7nN+Bo9wwnNppjdnsOPUbus8o1a9vY1QIDAQABAoIBAQDl1SBY3PlN36SF
yScUtCALdUbi4taVxkVxBbioQlFIKHGGkRD9JN/dgSApK6r36FdXNhAi40cQ4nnZ
iqd8FKqTSTFNa/mPM9ee+ITMI8nwOz8SiYcKTndPF2/yzapXDYDgCFcpz/czQ2X2
/i+IFyA5k4dUVomVGhFLBZ71xW5BvGUBMUH0XkeR5+c4gLvgR209BlpBHlkX4tUQ
+RQoxbKpkntl0mjqf91zcOe4LJVsXZFyN+NVSzLEbGC3lVSSiyjVQH3s7ExnTaHi
PpwSoXzu5QJj5xRit/1B3/LEGpIlPGFrkhMzBDTN+HYV/VLbCHJzjg5GVJawA82E
h2BY6YWJAoGBAPmGaZL5ggnTVR2XVBLDKbwL/sesqiPZk45B+I5eObHl+v236JH9
RPMjdE10jOR1TzfQdmE2/RboKhiVn+osS+2W6VXSo7sMsSM1bLBPYhnwrNIqzrX8
Vgi2bCl2S8ZhVo2R8c5WUaD0Gpxs6hwPIMOQWWwxDlsbg/UoLrhD3X4XAoGBAPFg
VSvaWQdDVAqjM42ObhZtWxeLfEAcxRQDMQq7btrTwBZSrtP3S3Egu66cp/4PT4VD
Hc8tYyT2rNETiqT6b2Rm1MgeoJ8wRqte6ZXSQVVQUOd42VG04O3aaleAGhXjEkM2
avctRdKHDhQdIt+riPgaNj4FdYpmQ5zIrcZtBr/zAoGBAOBXzBX7xMHmwxEe3NUd
qSlMM579C9+9oF/3ymzeJMtgtcBmGHEhoFtmVgvJrV8+ZaIOCFExam2tASQnaqbV
etK7q0ChaNok+CJqxzThupcN/6PaHw4aOJQOx8KjfE95dqNEQ367txqaPk7D0dy2
cUPDRdLzbC/X1lWV8iNzyPGzAoGBAN4R2epRpYz4Fa7/vWNkAcaib6c2zmaR0YN6
+Di+ftvW6yfehDhBkWgQTHv2ZtxoK6oYOKmuQUP1qsNkbi8gtTEzJlrDStWKbcom
tVMAsNkT3otHdPEmL7bFNwcvtVAjrF6oBztHrLBnTr2UnMwZnhdczkC7dwuQ0G3D
d5VSI16fAoGAY7eeVDkic73GbZmtZibuodvPJ/z85RIBOrzf3ColO4jGI6Ej/EnD
rMEe/mRC27CJzS9L9Jc0Kt66mGSvodDGl0nBsXGNfPog0cGwweCVN0Eo2VJZbRTT
UoU05/Pvu2h3/E8gGTBY0/WPSo06YUsICjVDWNuOIa/7IY7SyE6Xxn0=
-----END RSA PRIVATE KEY-----`

const exampleECP256PrivateKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEID1fSsJ8qyEqj2DVkrshaNiXqaSDX7qViASRkyGGJFbEoAoGCCqGSM49
AwEHoUQDQgAENnBG+bBJIaIa+bRlHaLiXD86RAy+Ef9CVdAfpPGoNRfkOTcrrIV7
2wv3Y5e0he63Tn9iVAFYRFexK1mjFw7TfA==
-----END EC PRIVATE KEY-----`

const exampleECP384PrivateKey = `
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDBPoOfapxtgZ8XNE7Wwdlw+9oDc6x4m57MITZyWzN62jkFUAYsvPJDF
9+g+e8CT5yqgBwYFK4EEACKhZANiAAQ2uZ0HIIxIavyjGyX13tIZVOaRB4+D64dF
s3DXDrpXcuDTSohw9xBW5sLDqRVu2LkBsCUFXtEJUHgC+O7wToNw8nh+KdDrcu/J
miNqbvEHuvlSlHWyx9HH8kAEuu1+SZg=
-----END EC PRIVATE KEY-----`

const exampleECP521PrivateKey = `
-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIBT07AnitDd1Z01bl5W5VW8/vTWyu7w3MSqEmCeKcM19p/TAJAeS8L
6UOig2fTUeuMeA2PoOUjI2Bid927VsWcxE2gBwYFK4EEACOhgYkDgYYABAGnV9mu
xY0E7/k8b+glOOMaN0+Qt70H9OmSz6tC8tU3EayRwFlNPch9TlvEpbCS3MsDE9dN
78EpFx45MUqzzdZcOgAu+EUC9Zas1YVK+WMo0GFy+XtFq3kxubOclBb52M/63mcd
zZnA8aAu9iTK9YPfcw1YWTJliNdKUoxmGVV5Ca1W4w==
-----END EC PRIVATE KEY-----`
