package oidc_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldNotRaiseErrorOnEqualPasswordsPlainText(t *testing.T) {
	hasher, err := oidc.NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc")
	b := []byte("abc")

	ctx := context.TODO()

	assert.NoError(t, hasher.Compare(ctx, a, b))
}

func TestShouldNotRaiseErrorOnEqualPasswordsPlainTextWithSeparator(t *testing.T) {
	hasher, err := oidc.NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc$123")
	b := []byte("abc$123")

	ctx := context.TODO()

	assert.NoError(t, hasher.Compare(ctx, a, b))
}

func TestShouldRaiseErrorOnNonEqualPasswordsPlainText(t *testing.T) {
	hasher, err := oidc.NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc")
	b := []byte("abcd")

	ctx := context.TODO()

	assert.EqualError(t, hasher.Compare(ctx, a, b), "The provided client secret did not match the registered client secret.")
}

func TestShouldHashPassword(t *testing.T) {
	hasher := oidc.Hasher{}

	data := []byte("abc")

	ctx := context.TODO()

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

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

func (s *ClientAuthenticationStrategySuite) GetCtx() oidc.Context {
	return &TestContext{
		Context:       context.TODO(),
		MockIssuerURL: s.GetIssuerURL(),
		Clock:         &utils.RealClock{},
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
				ID:                  "hs256",
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA256,
			},
			{
				ID:                  "hs384",
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA384,
			},
			{
				ID:                  "hs512",
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA512,
			},
			{
				ID:                  rs256,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA256,
			},
			{
				ID:                  "rs384",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA384,
			},
			{
				ID:                  "rs512",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA512,
			},
			{
				ID:                  "ps256",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA256,
			},
			{
				ID:                  "ps384",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA384,
			},
			{
				ID:                  "ps512",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA512,
			},
			{
				ID:                  "es256",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP256AndSHA256,
			},
			{
				ID:                  "es384",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP384AndSHA384,
			},
			{
				ID:                  es512,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP521AndSHA512,
			},

			{
				ID:                  "rs256k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: rs256, Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "rs384k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA384,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "rs384", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAUsingSHA384, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "rs512k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA512,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "rs512", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAUsingSHA512, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "ps256k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "ps256", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "ps384k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA384,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "ps384", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAPSSUsingSHA384, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "ps512k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAPSSUsingSHA512,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "ps512", Key: keyRSA2048.PublicKey, Algorithm: oidc.SigningAlgRSAPSSUsingSHA512, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "es256k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP256AndSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "es256", Key: keyECDSAP256.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "es384k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP384AndSHA384,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: "es384", Key: keyECDSAP384.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP384AndSHA384, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "es512k",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP521AndSHA512,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: es512, Key: keyECDSAP521.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "mismatched-alg",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: es512, Key: keyECDSAP521.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512, Use: oidc.KeyUseSignature},
					},
				},
			},
			{
				ID:                  "no-key",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgRSAUsingSHA256,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{},
				},
			},
			{
				ID:                  "es512u",
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodPrivateKeyJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgECDSAUsingP521AndSHA512,
				PublicKeys: schema.OpenIDConnectClientPublicKeys{
					Values: []schema.JWK{
						{KeyID: es512, Key: keyECDSAP521.PublicKey, Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512, Use: "enc"},
					},
				},
			},
			{
				ID:                  "hs5122",
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA512,
			},
			{
				ID:                  "hashed",
				Secret:              MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng"),
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA512,
			},
			{
				ID:                  oidc.ClientAuthMethodClientSecretBasic,
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretBasic,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
			},
			{
				ID:                  oidc.ClientAuthMethodNone,
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodNone,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
			},
			{
				ID:                  oidc.ClientAuthMethodClientSecretPost,
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretPost,
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
			},
			{
				ID:                  "bad_method",
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
				TokenEndpointAuthMethod:     "bad_method",
				TokenEndpointAuthSigningAlg: oidc.SigningAlgNone,
			},
			{
				ID:                  "base",
				Secret:              secret,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:                      "public",
				Public:                  true,
				AuthorizationPolicy:     authorization.OneFactor.String(),
				TokenEndpointAuthMethod: oidc.ClientAuthMethodNone,
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:                  "public-nomethod",
				Public:              true,
				AuthorizationPolicy: authorization.OneFactor.String(),
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:                      "public-basic",
				Public:                  true,
				AuthorizationPolicy:     authorization.OneFactor.String(),
				TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretBasic,
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:                      "public-post",
				Public:                  true,
				AuthorizationPolicy:     authorization.OneFactor.String(),
				TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
			{
				ID:                      "confidential-none",
				AuthorizationPolicy:     authorization.OneFactor.String(),
				TokenEndpointAuthMethod: oidc.ClientAuthMethodNone,
				RedirectURIs: []string{
					"https://client.example.com",
				},
			},
		},
	}, s.store, nil)

	c, _ := s.provider.Store.GetFullClient(context.TODO(), "no-key")
	client := c.(*oidc.FullClient)

	client.SetJSONWebKeys(&jose.JSONWebKeySet{})
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateAssertionHS256() {
	assertion := NewAssertion("hs256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal("hs256", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateAssertionHS384() {
	assertion := NewAssertion("hs384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS384, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal("hs384", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateAssertionHS512() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal("hs512", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnMismatchedAssertionAuthMethodPrivateKeyJWT() {
	assertion := NewAssertion(rs256, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client supports client authentication method 'private_key_jwt', however the 'HS512' JWA is not supported with this method.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnInvalidIssuerValue() {
	assertion := NewAssertionMapClaims("hs512", 123, s.GetTokenURL(), &jwt.NumericDate{Time: time.Now().Add(time.Second * -3)}, &jwt.NumericDate{Time: time.Unix(time.Now().Add(time.Minute).Unix(), 0)})

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'iss' from 'client_assertion' is invalid.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnInvalidSubjectValue() {
	assertion := NewAssertionMapClaims(123, "hs512", s.GetTokenURL(), &jwt.NumericDate{Time: time.Now().Add(time.Second * -3)}, &jwt.NumericDate{Time: time.Unix(time.Now().Add(time.Minute).Unix(), 0)})

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'sub' from 'client_assertion' is invalid.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnInvalidJTIValue() {
	assertion := NewAssertionMapClaims("hs512", "hs512", s.GetTokenURL(), &jwt.NumericDate{Time: time.Now().Add(time.Second * -3)}, &jwt.NumericDate{Time: time.Unix(time.Now().Add(time.Minute).Unix(), 0)})

	assertion[oidc.ClaimJWTID] = 123

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'jti' from 'client_assertion' is invalid.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnMismatchedAssertionAuthMethodClientSecretJWT() {
	assertion := NewAssertion("hs256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client supports client authentication method 'client_secret_jwt', however the 'PS256' JWA is not supported with this method.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnMismatchedAlg() {
	assertion := NewAssertion(rs256, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'PS256' but the requested OAuth 2.0 Client enforces signing algorithm 'RS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnClientAssertionWithClientSecret() {
	assertion := NewAssertion(rs256, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)
	r.PostForm.Set(oidc.FormParameterClientSecret, "abc123")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The client_secret request parameter must not be set when using client_assertion_type of 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnClientAssertionWithAuthorizationHeader() {
	assertion := NewAssertion(rs256, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)
	r.Header.Set(fasthttp.HeaderAuthorization, "abc123")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The Authorization request header must not be set when using client_assertion_type of 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnMismatchedAlgSameMethod() {
	assertion := NewAssertion("hs256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'HS512' but the requested OAuth 2.0 Client enforces signing algorithm 'HS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS256() {
	assertion := NewAssertion(rs256, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnBadAlgRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = rs256

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'PS256' but the requested OAuth 2.0 Client enforces signing algorithm 'RS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnBadKidRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "nokey"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The JSON Web Token uses signing key with kid 'nokey', which could not be found.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnBadTypRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = rs256

	token, err := assertionJWT.SignedString(keyECDSAP256)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The 'client_assertion' uses signing algorithm 'ES256' but the requested OAuth 2.0 Client enforces signing algorithm 'RS256'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysRS256() {
	assertion := NewAssertion("rs256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = rs256

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("rs256k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS384() {
	assertion := NewAssertion("rs384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS384, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysRS384() {
	assertion := NewAssertion("rs384k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS384, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "rs384"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("rs384k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysRS512() {
	assertion := NewAssertion("rs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS512, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysRS512() {
	assertion := NewAssertion("rs512k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS512, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "rs512"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("rs512k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS256() {
	assertion := NewAssertion("ps256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysPS256() {
	assertion := NewAssertion("ps256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "ps256"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("ps256k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS384() {
	assertion := NewAssertion("ps384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS384, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysPS384() {
	assertion := NewAssertion("ps384k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS384, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "ps384"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("ps384k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysPS512() {
	assertion := NewAssertion("ps512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS512, assertion)

	token, err := assertionJWT.SignedString(keyRSA2048)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysPS512() {
	assertion := NewAssertion("ps512k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodPS512, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "ps512"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("ps512k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES256() {
	assertion := NewAssertion("es256", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES256, assertion)

	token, err := assertionJWT.SignedString(keyECDSAP256)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysES256() {
	assertion := NewAssertion("es256k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "es256"

	token, err := assertionJWT.SignedString(keyECDSAP256)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("es256k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES384() {
	assertion := NewAssertion("es384", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES384, assertion)

	token, err := assertionJWT.SignedString(keyECDSAP384)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysES384() {
	assertion := NewAssertion("es384k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES384, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "es384"

	token, err := assertionJWT.SignedString(keyECDSAP384)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("es384k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnUnregisteredKeysES512() {
	assertion := NewAssertion(es512, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES512, assertion)

	token, err := assertionJWT.SignedString(keyECDSAP521)

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client has no JSON Web Keys set registered, but they are needed to complete the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldAuthKeysES512() {
	assertion := NewAssertion("es512k", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES512, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = es512

	token, err := assertionJWT.SignedString(keyECDSAP521)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)

	s.Equal("es512k", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailKeysES512Enc() {
	assertion := NewAssertion("es512u", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodES512, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = es512

	token, err := assertionJWT.SignedString(keyECDSAP521)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Unable to find ECDSA public key with a 'use' value of 'sig' for kid 'es512' in JSON Web Key Set.")
	s.Require().Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailKeysRS256MismatchedKIDAndAlg() {
	assertion := NewAssertion("mismatched-alg", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = es512

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Unable to find RSA public key with a 'use' value of 'sig' for kid 'es512' in JSON Web Key Set.")
	s.Require().Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithoutKeys() {
	assertion := NewAssertion("no-key", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, assertion)
	assertionJWT.Header[oidc.JWTHeaderKeyIdentifier] = "rs256"

	token, err := assertionJWT.SignedString(keyRSA2048)
	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The retrieved JSON Web Key Set does not contain any key.")
	s.Require().Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnJTIKnown() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The jti was already used.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateJWTWithArbitraryClaims() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	a["aaa"] = "abc"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal("hs512", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingSubClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	delete(a, oidc.ClaimSubject)

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'sub' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client. The claim 'sub' with value '' did not match the 'client_id' with value 'hs512'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithInvalidExpClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	a[oidc.ClaimExpirationTime] = "not a number"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token claims are invalid.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingIssClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	delete(a, oidc.ClaimIssuer)

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'iss' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client. The claim 'iss' with value '' did not match the 'client_id' with value 'hs512'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithInvalidAudClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertion.Audience = []string{"notvalid"}

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
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

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'aud' from 'client_assertion' must match the authorization server's token endpoint 'https://auth.example.com/api/oidc/token'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithInvalidAssertionType() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	v := s.GetAssertionValues(token)

	v.Set(oidc.FormParameterClientAssertionType, "not_valid")

	r := s.GetRequest(v)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Unknown client_assertion_type 'not_valid'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingJTIClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	a := assertion.ToMapClaims()
	delete(a, oidc.ClaimJWTID)

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, a)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'jti' from 'client_assertion' must be set but it is not.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMismatchedIssClaim() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertion.Issuer = "hs256"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'iss' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client. The claim 'iss' with value 'hs256' did not match the 'client_id' with value 'hs512'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateClientSecretPost() {
	r := s.GetClientSecretPostRequest(oidc.ClientAuthMethodClientSecretPost, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal(oidc.ClientAuthMethodClientSecretPost, client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretPostOnClientSecretBasicClient() {
	r := s.GetClientSecretPostRequest(oidc.ClientAuthMethodClientSecretBasic, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client supports client authentication method 'client_secret_basic', but method 'client_secret_post' was requested. You must configure the OAuth 2.0 client's 'token_endpoint_auth_method' value to accept 'client_secret_post'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretPostWrongSecret() {
	r := s.GetClientSecretPostRequest(oidc.ClientAuthMethodClientSecretPost, "client-secret-bad")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The provided client secret did not match the registered client secret.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidateClientSecretBasic() {
	r := s.GetClientSecretBasicRequest(oidc.ClientAuthMethodClientSecretBasic, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal(oidc.ClientAuthMethodClientSecretBasic, client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldRaiseErrorOnClientSecretPostWithoutClientID() {
	r := s.GetRequest(&url.Values{oidc.FormParameterClientSecret: []string{"client-secret"}})

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Client credentials missing or malformed in both HTTP Authorization header and HTTP POST body.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldHandleBasicAuth() {
	r := s.GetRequest(&url.Values{oidc.FormParameterRequestURI: []string{"not applicable"}})

	r.Header.Set(fasthttp.HeaderAuthorization, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("client_secret_basic:client-secret"))))
	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal("client_secret_basic", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorWithBasicAuthBadBas64Data() {
	r := s.GetRequest(&url.Values{oidc.FormParameterRequestURI: []string{"not applicable"}})

	r.Header.Set(fasthttp.HeaderAuthorization, fmt.Sprintf("Basic %s", "!@#&!@*&#(!*@#(*!@#"))
	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorWithBasicAuthBadBasicData() {
	r := s.GetRequest(&url.Values{oidc.FormParameterRequestURI: []string{"not applicable"}})

	r.Header.Set(fasthttp.HeaderAuthorization, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("client_secret_basic"))))
	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorWithBasicAuthNoScheme() {
	r := s.GetRequest(&url.Values{oidc.FormParameterRequestURI: []string{"not applicable"}})

	r.Header.Set(fasthttp.HeaderAuthorization, fmt.Sprintf("Basic%s", base64.StdEncoding.EncodeToString([]byte("client_secret_basic:client-secret"))))
	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorWithBasicAuthInvalidScheme() {
	r := s.GetRequest(&url.Values{oidc.FormParameterRequestURI: []string{"not applicable"}})

	r.Header.Set(fasthttp.HeaderAuthorization, fmt.Sprintf("Bassic %s", base64.StdEncoding.EncodeToString([]byte("client_secret_basic:client-secret"))))
	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnClientSecretPostClient() {
	r := s.GetClientSecretBasicRequest(oidc.ClientAuthMethodClientSecretPost, "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client supports client authentication method 'client_secret_post', but method 'client_secret_basic' was requested. You must configure the OAuth 2.0 client's 'token_endpoint_auth_method' value to accept 'client_secret_basic'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicWrongSecret() {
	r := s.GetClientSecretBasicRequest(oidc.ClientAuthMethodClientSecretBasic, "client-secret-bad")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The provided client secret did not match the registered client secret.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnPublic() {
	r := s.GetClientSecretBasicRequest("public", "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client supports client authentication method 'none', but method 'client_secret_basic' was requested. You must configure the OAuth 2.0 client's 'token_endpoint_auth_method' value to accept 'client_secret_basic'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnPublicWithBasic() {
	r := s.GetClientSecretBasicRequest("public-basic", "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client is not a confidential client however the client authentication method 'client_secret_basic' was used which is not permitted as it's only permitted on confidential clients.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretPostOnPublicWithPost() {
	r := s.GetClientSecretPostRequest("public-post", "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client is not a confidential client however the client authentication method 'client_secret_post' was used which is not permitted as it's only permitted on confidential clients.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorNoneOnConfidentialWithNone() {
	r := s.GetClientSecretPostRequest("confidential-none", "")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). The OAuth 2.0 Client is a confidential client however the client authentication method 'none' was used which is not permitted as it's not permitted on confidential clients.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldErrorClientSecretBasicOnInvalidClient() {
	r := s.GetClientSecretBasicRequest("not-a-client", "client-secret")

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Client with id 'not-a-client' does not appear to be a registered client.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldValidatePublic() {
	v := s.GetClientValues("public")

	r := s.GetRequest(v)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotNil(client)
	s.Equal("public", client.GetID())
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMismatchedFormClientID() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertion.Issuer = "hs5122"

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	values := s.GetAssertionValues(token)

	values.Set(oidc.FormParameterClientID, "hs5122")

	r := s.GetRequest(values)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'sub' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client. The claim 'sub' with value 'hs512' did not match the 'client_id' with value 'hs5122'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMismatchedFormClientIDWithIss() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	values := s.GetAssertionValues(token)

	values.Set(oidc.FormParameterClientID, "hs5122")

	r := s.GetRequest(values)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Claim 'iss' from 'client_assertion' must match the 'client_id' of the OAuth 2.0 Client. The claim 'iss' with value 'hs512' did not match the 'client_id' with value 'hs5122'.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailWithMissingClient() {
	assertion := NewAssertion("noclient", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	client, err := s.provider.DefaultClientAuthenticationStrategy(s.GetCtx(), r, r.PostForm)

	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Client with id 'noclient' does not appear to be a registered client.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailBadSecret() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret-wrong"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The signature is invalid.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailMethodNone() {
	assertion := NewAssertion(oidc.ClientAuthMethodNone, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client does not support client authentication, however 'client_assertion' was provided in the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionMethodClientSecretPost() {
	assertion := NewAssertion(oidc.ClientAuthMethodClientSecretPost, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client only supports client authentication method 'client_secret_post', however 'client_assertion' was provided in the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionMethodBad() {
	assertion := NewAssertion("bad_method", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client only supports client authentication method 'bad_method', however that method is not supported by this server.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionBaseClient() {
	assertion := NewAssertion("base", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The client configuration does not support OpenID Connect specific authentication methods.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailAssertionMethodClientSecretBasic() {
	assertion := NewAssertion(oidc.ClientAuthMethodClientSecretBasic, s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This requested OAuth 2.0 client only supports client authentication method 'client_secret_basic', however 'client_assertion' was provided in the request.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailHashedSecret() {
	assertion := NewAssertion("hashed", s.GetTokenURL(), time.Now().Add(time.Second*-3), time.Unix(time.Now().Add(time.Minute).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). This client does not support authentication method 'client_secret_jwt' as the client secret is not in plaintext.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailExpiredToken() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Minute*-3), time.Unix(time.Now().Add(time.Minute*-1).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token is expired.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailNotYetValid() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Minute*-3), time.Unix(time.Now().Add(time.Minute*1).Unix(), 0))

	assertion.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Second * 10))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	s.NoError(r.ParseForm())

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token isn't valid yet.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailTokenUsedBeforeIssued() {
	assertion := NewAssertion("hs512", s.GetTokenURL(), time.Now().Add(time.Minute*3), time.Unix(time.Now().Add(time.Minute*8).Unix(), 0))

	assertionJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, assertion)

	token, err := assertionJWT.SignedString([]byte("client-secret"))

	s.Require().NoError(oidc.ErrorToDebugRFC6749Error(err))
	s.Require().NotEqual("", token)

	r := s.GetAssertionRequest(token)

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token was used before it was issued.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailMalformed() {
	r := s.GetAssertionRequest("bad token")

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_client")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Unable to verify the integrity of the 'client_assertion' value. The token is malformed.")
	s.Nil(client)
}

func (s *ClientAuthenticationStrategySuite) TestShouldFailMissingAssertion() {
	r := s.GetAssertionRequest("")

	ctx := s.GetCtx()

	client, err := s.provider.DefaultClientAuthenticationStrategy(ctx, r, r.PostForm)

	s.EqualError(err, "invalid_request")
	s.EqualError(oidc.ErrorToDebugRFC6749Error(err), "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The 'client_assertion' request parameter must be set when using 'client_assertion_type' of 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer'.")
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

func NewAssertionMapClaims(subject, issuer any, tokenURL *url.URL, iat, exp any) jwt.MapClaims {
	return jwt.MapClaims{
		oidc.ClaimJWTID:          uuid.Must(uuid.NewRandom()).String(),
		oidc.ClaimIssuer:         issuer,
		oidc.ClaimSubject:        subject,
		oidc.ClaimIssuedAt:       iat,
		oidc.ClaimExpirationTime: exp,
		oidc.ClaimAudience: []string{
			tokenURL.String(),
		},
	}
}

func TestPKCEHandler_HandleAuthorizeEndpointRequest_Mock(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockPKCERequestStorage(ctrl)

	defer ctrl.Finish()

	strategy := &TestCodeStrategy{}
	config := &oidc.Config{
		ProofKeyCodeExchange: oidc.ProofKeyCodeExchangeConfig{
			AllowPlainChallengeMethod: true,
		},
	}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	request := &fosite.AuthorizeRequest{
		ResponseTypes: fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
		Request: fosite.Request{
			Client: &fosite.DefaultClient{ID: "test"},
			Form: url.Values{
				oidc.FormParameterCodeChallenge: []string{pkceVerifier},
			},
		},
	}

	response := fosite.NewAuthorizeResponse()

	response.AddParameter(oidc.FormParameterAuthorizationCode, "abc123")

	ctx := context.Background()

	store.EXPECT().CreatePKCERequestSession(ctx, gomock.Any(), gomock.Any()).Return(errors.New("bad storage error!"))

	assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(handler.HandleAuthorizeEndpointRequest(ctx, request, response)), "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. bad storage error!")
}

func TestPKCEHandler_Misc(t *testing.T) {
	store := storage.NewMemoryStore()
	strategy := &TestCodeStrategy{}
	config := &oidc.Config{}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	assert.Nil(t, handler.PopulateTokenEndpointResponse(context.TODO(), nil, nil))
	assert.False(t, handler.CanSkipClientAuth(context.TODO(), nil))
}

func TestPKCEHandler_CanHandleTokenEndpointRequest(t *testing.T) {
	testCases := []struct {
		name     string
		have     fosite.AccessRequester
		expected bool
	}{
		{
			"ShouldHandleAuthorizeCode",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeAuthorizationCode},
			},
			true,
		},
		{
			"ShouldNotHandleRefreshToken",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
			},
			false,
		},
		{
			"ShouldNotHandleClientCredentials",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
			},
			false,
		},
		{
			"ShouldNotHandleImplicit",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeImplicit},
			},
			false,
		},
	}

	store := storage.NewMemoryStore()
	strategy := &TestCodeStrategy{}
	config := &oidc.Config{}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, handler.CanHandleTokenEndpointRequest(context.TODO(), tc.have))
		})
	}
}

func TestPKCEHandler_HandleAuthorizeEndpointRequest(t *testing.T) {
	store := storage.NewMemoryStore()
	strategy := &TestCodeStrategy{}
	config := &oidc.Config{}

	client := &oidc.BaseClient{ID: "test"}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	testCases := []struct {
		name                                      string
		types                                     fosite.Arguments
		enforce, enforcePublicClients, allowPlain bool
		method, challenge, code                   string
		expected                                  string
		client                                    *oidc.BaseClient
	}{
		{
			"ShouldNotHandleBlankResponseModes",
			fosite.Arguments{},
			false,
			false,
			false,
			oidc.PKCEChallengeMethodPlain,
			"challenge",
			"",
			"",
			client,
		},
		{
			"ShouldHandleAuthorizeCodeFlow",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			false,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"challenge",
			"abc123",
			"",
			client,
		},
		{
			"ShouldErrorHandleAuthorizeCodeFlowWithoutCode",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			false,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"challenge",
			"",
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. The PKCE handler must be loaded after the authorize code handler.",
			client,
		},
		{
			"ShouldErrorHandleAuthorizeCodeFlowWithoutChallengeMethodWhenEnforced",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			true,
			false,
			true,
			"",
			"",
			"abc123",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for clients.",
			client,
		},
		{
			"ShouldErrorHandleAuthorizeCodeFlowWithoutChallengeWhenEnforced",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"",
			"abc123",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for clients.",
			client,
		},
		{
			"ShouldSkipNotEnforced",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			false,
			false,
			true,
			"",
			"",
			"abc123",
			"",
			client,
		},
		{
			"ShouldErrorUnknownChallengeMethod",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			true,
			false,
			true,
			"abc",
			"abc",
			"abc123",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The code_challenge_method is not supported, use S256 instead.",
			client,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config.ProofKeyCodeExchange.Enforce = tc.enforce
			config.ProofKeyCodeExchange.EnforcePublicClients = tc.enforcePublicClients
			config.ProofKeyCodeExchange.AllowPlainChallengeMethod = tc.allowPlain

			requester := fosite.NewAuthorizeRequest()

			requester.Client = tc.client

			if len(tc.method) > 0 {
				requester.Form.Add(oidc.FormParameterCodeChallengeMethod, tc.method)
			}

			if len(tc.challenge) > 0 {
				requester.Form.Add(oidc.FormParameterCodeChallenge, tc.challenge)
			}

			requester.ResponseTypes = tc.types

			responder := fosite.NewAuthorizeResponse()

			if len(tc.code) > 0 {
				responder.AddParameter(oidc.FormParameterAuthorizationCode, tc.code)
			}

			err := handler.HandleAuthorizeEndpointRequest(context.TODO(), requester, responder)

			err = oidc.ErrorToDebugRFC6749Error(err)

			if len(tc.expected) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestPKCEHandler_HandleTokenEndpointRequest(t *testing.T) {
	store := storage.NewMemoryStore()
	strategy := &TestCodeStrategy{}
	config := &oidc.Config{}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	clientConfidential := &oidc.BaseClient{
		ID:     "test",
		Public: false,
	}

	clientPublic := &oidc.BaseClient{
		ID:     "test",
		Public: true,
	}

	testCases := []struct {
		name                                      string
		grant                                     string
		enforce, enforcePublicClients, allowPlain bool
		method, challenge, verifier               string
		code                                      string
		expected                                  string
		client                                    *oidc.BaseClient
	}{
		{
			"ShouldFailNotAuthCode",
			oidc.GrantTypeClientCredentials,
			false,
			false,
			false,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			pkceVerifier,
			"code-0",
			"The handler is not responsible for this request.",
			clientConfidential,
		},
		{
			"ShouldPassPlainWithConfidentialClientWhenEnforcedWhenAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodPlain,
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"code-1",
			"",
			clientConfidential,
		},
		{
			"ShouldFailWithConfidentialClientWithNotAllowPlainWhenPlainWhenEnforced",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			false,
			oidc.PKCEChallengeMethodPlain,
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"code-2",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must use the 'S256' PKCE 'code_challenge_method' but the 'plain' method was requested. The server is configured in a way that enforces PKCE 'S256' as challenge method for clients.",
			clientConfidential,
		},
		{
			"ShouldPassWithConfidentialClientWhenNotProvidedWhenNotEnforced",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			false,
			"",
			"",
			"",
			"code-3",
			"",
			clientConfidential,
		},
		{
			"ShouldPassWithPublicClientWhenNotProvidedWhenNotEnforced",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			false,
			"",
			"",
			"",
			"code-4",
			"",
			clientPublic,
		},
		{
			"ShouldFailWithPublicClientWhenNotProvidedWhenNotEnforcedWhenEnforcedForPublicClients",
			oidc.GrantTypeAuthorizationCode,
			false,
			true,
			false,
			"",
			"",
			"",
			"code-5",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. This client must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for this client.",
			clientPublic,
		},
		{
			"ShouldPassS256WithConfidentialClientWhenEnforcedWhenAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			pkceVerifier,
			"code-6",
			"",
			clientConfidential,
		},
		{
			"ShouldPassS256WithConfidentialClientWhenEnforcedWhenNotAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			false,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			pkceVerifier,
			"code-7",
			"",
			clientConfidential,
		},
		{
			"ShouldFailS256WithConfidentialClientWhenEnforcedWhenAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			"jSfm3f5nS9eD4eYSHaeQxVBVKxXnfmbWAFQiiAdMAK98EhNifm",
			"code-8",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code challenge did not match the code verifier.",
			clientConfidential,
		},
		{
			"ShouldFailS256WithConfidentialClientWhenVerifierTooShort",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			"aaaaaaaaaaaaa",
			"code-9",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier must be at least 43 characters.",
			clientConfidential,
		},
		{
			"ShouldFailPlainWithConfidentialClientWhenNotMatching",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodPlain,
			"jqNhtkWBNbP9oz6BjA2ufPWxADqnHJcUNhS8VNzBWLd44ynFzi",
			"uoPgkXQNCiiMzQ4aXeXdzBQaDArGNN9bke8gQWo7qZZ2djrcJZ",
			"code-10",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code challenge did not match the code verifier.",
			clientConfidential,
		},
		{
			"ShouldPassNoPKCESessionWhenNotEnforced",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			true,
			"",
			"",
			"",
			"code-11",
			"",
			clientConfidential,
		},
		{
			"ShouldFailNoPKCESessionWhenEnforced",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			"",
			"",
			"",
			"code-12",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for clients.",
			clientConfidential,
		},
		{
			"ShouldFailLongVerifier",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"fUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJfUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJ",
			"fUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJfUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJ",
			"code-13",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier can not be longer than 128 characters.",
			clientConfidential,
		},
		{
			"ShouldFailBadChars",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"U5katcVxpTd8jU5YpUD9xdoivT55zzMWdMiy2TDA4DH9FJ5bTK45Mhd@",
			"U5katcVxpTd8jU5YpUD9xdoivT55zzMWdMiy2TDA4DH9FJ5bTK45Mhd@",
			"code-14",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier must only contain [a-Z], [0-9], '-', '.', '_', '~'.",
			clientConfidential,
		},
		{
			"ShouldFailNoChallenge",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"",
			"U5katcVxpTd8jU5YpUD9xdoivT55zzMWdMiy2TDA4DH9FJ5bTK45Mhd",
			"code-15",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier was provided but the code challenge was absent from the authorization request.",
			clientConfidential,
		},
		{
			"ShouldFailNoPKCESessionWhenEnforcedNotFound",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"",
			"abc123",
			"",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. Unable to find initial PKCE data tied to this request. Could not find the requested resource(s).",
			clientConfidential,
		},
		{
			"ShouldFailInvalidCode",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"WXNqfH6FCXcJH5oT9eqTM3HTdTh4b2aSvVe9KWkxcHCJJ3FaXF",
			"WXNqfH6FCXcJH5oT9eqTM3HTdTh4b2aSvVe9KWkxcHCJJ3FaXF",
			"BADCODE",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. Unable to find initial PKCE data tied to this request. Could not find the requested resource(s).",
			clientConfidential,
		},
		{
			"ShouldPassNotEnforcedNoSession",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			true,
			"",
			"",
			"",
			"",
			"",
			clientConfidential,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()

			config.ProofKeyCodeExchange.Enforce = tc.enforce
			config.ProofKeyCodeExchange.EnforcePublicClients = tc.enforcePublicClients
			config.ProofKeyCodeExchange.AllowPlainChallengeMethod = tc.allowPlain

			if len(tc.code) > 0 {
				strategy.signature = tc.code
			} else {
				strategy.signature = fmt.Sprintf("code-%d", i)
			}

			ar := fosite.NewAuthorizeRequest()

			ar.Client = tc.client

			if len(tc.challenge) != 0 {
				ar.Form.Add(oidc.FormParameterCodeChallenge, tc.challenge)
			}

			if len(tc.method) != 0 {
				ar.Form.Add(oidc.FormParameterCodeChallengeMethod, tc.method)
			}

			if len(tc.code) > 0 {
				require.NoError(t, store.CreatePKCERequestSession(ctx, fmt.Sprintf("code-%d", i), ar))
			}

			r := fosite.NewAccessRequest(nil)
			r.Client = tc.client
			r.GrantTypes = fosite.Arguments{tc.grant}

			if len(tc.verifier) != 0 {
				r.Form.Add(oidc.FormParameterCodeVerifier, tc.verifier)
			}

			err := handler.HandleTokenEndpointRequest(context.TODO(), r)
			err = oidc.ErrorToDebugRFC6749Error(err)

			if len(tc.expected) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestPKCEHandler_HandleTokenEndpointRequest_Mock(t *testing.T) {
	client := &oidc.BaseClient{
		ID:     "test",
		Public: false,
	}

	testCases := []struct {
		name                                      string
		setup                                     func(mock *mocks.MockPKCERequestStorage)
		grant                                     string
		enforce, enforcePublicClients, allowPlain bool
		method, challenge, verifier               string
		expected                                  string
		client                                    *oidc.BaseClient
	}{
		{
			"ShouldPassS256WithConfidentialClientWhenEnforcedWhenAllowPlain",
			func(mock *mocks.MockPKCERequestStorage) {
				mock.EXPECT().GetPKCERequestSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fosite.ErrNotFound.WithDebug("A bad error!"))
			},
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			pkceVerifier,
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. Unable to find initial PKCE data tied to this request. Could not find the requested resource(s). A bad error!",
			client,
		},
		{
			"ShouldPassS256WithConfidentialClientWhenEnforcedWhenAllowPlain",
			func(mock *mocks.MockPKCERequestStorage) {
				mock.EXPECT().GetPKCERequestSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("A really bad error!"))
			},
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			pkceVerifier,
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. A really bad error!",
			client,
		},
		{
			"ShouldPassS256WithConfidentialClientWhenEnforcedWhenAllowPlain",
			func(mock *mocks.MockPKCERequestStorage) {
				mock.EXPECT().GetPKCERequestSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.NewAccessRequest(&fosite.DefaultSession{}), nil)
				mock.EXPECT().DeletePKCERequestSession(gomock.Any(), gomock.Any()).Return(errors.New("Could not delete!"))
			},
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			pkceChallenge,
			pkceVerifier,
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Could not delete!",
			client,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mocks.NewMockPKCERequestStorage(ctrl)
			defer ctrl.Finish()

			strategy := &TestCodeStrategy{}
			config := &oidc.Config{
				ProofKeyCodeExchange: oidc.ProofKeyCodeExchangeConfig{
					Enforce:                   tc.enforce,
					AllowPlainChallengeMethod: tc.allowPlain,
					EnforcePublicClients:      tc.enforcePublicClients,
				},
			}

			handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

			if tc.setup != nil {
				tc.setup(store)
			}

			request := fosite.NewAccessRequest(&fosite.DefaultSession{})

			request.GrantTypes = fosite.Arguments{oidc.GrantTypeAuthorizationCode}
			request.Form = url.Values{
				oidc.FormParameterCodeVerifier: []string{tc.verifier},
			}

			err := oidc.ErrorToDebugRFC6749Error(handler.HandleTokenEndpointRequest(context.TODO(), request))

			if len(tc.expected) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

const (
	pkceChallenge = "GM6jKJIR6JxgxU5m5Y79WzudqoNmo7PogrhI1_F8eGw"
	pkceVerifier  = "Nt6MpT7QXtfme55cKv9b23KEAvSEHyjRVtQt5jcjUUmWU9bTzd"
)
