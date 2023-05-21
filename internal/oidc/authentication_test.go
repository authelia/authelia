package oidc_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestShouldNotRaiseErrorOnEqualPasswordsPlainText(t *testing.T) {
	hasher, err := oidc.NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc")
	b := []byte("abc")

	ctx := context.Background()

	assert.NoError(t, hasher.Compare(ctx, a, b))
}

func TestShouldNotRaiseErrorOnEqualPasswordsPlainTextWithSeparator(t *testing.T) {
	hasher, err := oidc.NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc$123")
	b := []byte("abc$123")

	ctx := context.Background()

	assert.NoError(t, hasher.Compare(ctx, a, b))
}

func TestShouldRaiseErrorOnNonEqualPasswordsPlainText(t *testing.T) {
	hasher, err := oidc.NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc")
	b := []byte("abcd")

	ctx := context.Background()

	assert.EqualError(t, hasher.Compare(ctx, a, b), "The provided client secret did not match the registered client secret.")
}

func TestShouldHashPassword(t *testing.T) {
	hasher := oidc.Hasher{}

	data := []byte("abc")

	ctx := context.Background()

	hash, err := hasher.Hash(ctx, data)

	assert.NoError(t, err)
	assert.Equal(t, data, hash)
}

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
	fmt.Println(s.GetIssuerURL())

	return &MockOpenIDConnectContext{
		Context:       context.Background(),
		MockIssuerURL: s.GetIssuerURL(),
	}
}

func (s *ClientAuthenticationStrategySuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mocks.NewMockStorage(s.ctrl)

	secret := tOpenIDConnectPlainTextClientSecret

	s.provider = oidc.NewOpenIDConnectProvider(&schema.OpenIDConnect{
		IssuerPrivateKeys: []schema.JWK{
			{Key: keyRSA2048, CertificateChain: certRSA2048, Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256},
		},
		HMACSecret: "abc123",
		Clients: []schema.OpenIDConnectClient{
			{
				ID:     "hs256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA256,
			},
			{
				ID:     "hs384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA384,
			},
			{
				ID:     "hs512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA512,
			},
			{
				ID:     rs256,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA256,
			},
			{
				ID:     "rs384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA384,
			},
			{
				ID:     "rs512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA512,
			},
			{
				ID:     "ps256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA256,
			},
			{
				ID:     "ps384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA384,
			},
			{
				ID:     "ps512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA512,
			},
			{
				ID:     "es256",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP256AndSHA256,
			},
			{
				ID:     "es384",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP384AndSHA384,
			},
			{
				ID:     "es512",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP521AndSHA512,
			},

			{
				ID:     "rs256k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: rs256, Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "rs384k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA384,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "rs384", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAUsingSHA384, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "rs512k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA512,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "rs512", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAUsingSHA512, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "ps256k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "ps256", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "ps384k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA384,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "ps384", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAPSSUsingSHA384, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "ps512k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA512,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "ps512", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAPSSUsingSHA512, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "es256k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP256AndSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "es256", Key: keyECDSAP256.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "es384k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP384AndSHA384,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "es384", Key: keyECDSAP384.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP384AndSHA384, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "es512k",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP521AndSHA512,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "es512", Key: keyECDSAP521.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:     "hs5122",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA512,
			},
			{
				ID:     "hashed",
				Secret: MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng"),
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA512,
			},
			{
				ID:     oidc.ClientAuthMethodClientSecretBasic,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretBasic,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
			},
			{
				ID:     oidc.ClientAuthMethodNone,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodNone,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
			},
			{
				ID:     oidc.ClientAuthMethodClientSecretPost,
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretPost,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
			},
			{
				ID:     "bad_method",
				Secret: secret,
				Policy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     "bad_method",
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
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
	assertion := NewAssertion(rs256, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

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
	assertion := NewAssertion(rs256, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnBadAlgRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = rs256

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'PS256' but the requested OAuth 2.0 Client enforces signing algorithm 'RS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnBadKidRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "nokey"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The JSON Web Token uses signing key with kid 'nokey', which could not be found.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnBadTypRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = rs256

	token, err := assertionJWT.SignedString(keyECDSAP256)
	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'ES256' but the requested OAuth 2.0 Client enforces signing algorithm 'RS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = rs256

	token, err := assertionJWT.SignedString(keyRSA2048)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("rs256k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS384() {
	assertion := NewAssertion("rs384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS384, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysRS384() {
	assertion := NewAssertion("rs384k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS384, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "rs384"

	token, err := assertionJWT.SignedString(keyRSA2048)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("rs384k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS512() {
	assertion := NewAssertion("rs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS512, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysRS512() {
	assertion := NewAssertion("rs512k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS512, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "rs512"

	token, err := assertionJWT.SignedString(keyRSA2048)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("rs512k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS256() {
	assertion := NewAssertion("ps256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysPS256() {
	assertion := NewAssertion("ps256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "ps256"

	token, err := assertionJWT.SignedString(keyRSA2048)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("ps256k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS384() {
	assertion := NewAssertion("ps384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS384, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysPS384() {
	assertion := NewAssertion("ps384k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS384, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "ps384"

	token, err := assertionJWT.SignedString(keyRSA2048)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("ps384k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS512() {
	assertion := NewAssertion("ps512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS512, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysPS512() {
	assertion := NewAssertion("ps512k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS512, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "ps512"

	token, err := assertionJWT.SignedString(keyRSA2048)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("ps512k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES256() {
	assertion := NewAssertion("es256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES256, assertion)

	token, err := assertionJWT.SignedString(keyECDSAP256)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysES256() {
	assertion := NewAssertion("es256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "es256"

	token, err := assertionJWT.SignedString(keyECDSAP256)

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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("es256k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES384() {
	assertion := NewAssertion("es384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES384, assertion)

	token, err := assertionJWT.SignedString(keyECDSAP384)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysES384() {
	assertion := NewAssertion("es384k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES384, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "es384"

	token, err := assertionJWT.SignedString(keyECDSAP384)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("es384k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES512() {
	assertion := NewAssertion("es512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES512, assertion)

	token, err := assertionJWT.SignedString(keyECDSAP521)

	s.Require().NoError(err)
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(ErrorToRFC6749ErrorTest(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysES512() {
	assertion := NewAssertion("es512k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES512, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "es512"

	token, err := assertionJWT.SignedString(keyECDSAP521)
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

	s.NoError(err)
	s.Require().NotNil(client)

	s.Equal("es512k", client.GetID())
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
