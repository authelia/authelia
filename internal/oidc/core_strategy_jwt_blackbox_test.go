package oidc_test

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/token/hmac"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestJWTCoreStrategy(t *testing.T) {
	goodsecret := []byte("R7VCSUfnKc7Y5zE84q6GstYqfMGjL4wM")
	secreta := []byte("a")

	config := &oidc.Config{
		TokenEntropy: 10,
		GlobalSecret: secreta,
		Lifespans: schema.IdentityProvidersOpenIDConnectLifespanToken{
			AccessToken:   time.Hour,
			RefreshToken:  time.Hour,
			AuthorizeCode: time.Minute,
		},
	}

	strategy := &oidc.JWTCoreStrategy{
		Signer: &jwt.DefaultSigner{
			GetPrivateKey: func(ctx context.Context) (interface{}, error) {
				return x509PrivateKeyRSA2048, nil
			},
		},
		HMACCoreStrategy: &oidc.HMACCoreStrategy{
			Enigma: &hmac.HMACStrategy{Config: config},
			Config: config,
		},
		Config: config,
	}

	var (
		token, signature string
		err              error
	)

	ctx := context.Background()

	token, signature, err = strategy.GenerateAuthorizeCode(ctx, &fosite.Request{})
	assert.EqualError(t, err, "secret for signing HMAC-SHA512/256 is expected to be 32 byte long, got 1 byte")
	assert.Empty(t, token)
	assert.Empty(t, signature)

	config.GlobalSecret = goodsecret

	token, signature, err = strategy.GenerateAuthorizeCode(ctx, &fosite.Request{})
	assert.NoError(t, err)

	assert.NotEmpty(t, token)
	assert.NotEmpty(t, signature)
	assert.Equal(t, signature, strategy.AuthorizeCodeSignature(ctx, token))
	assert.Regexp(t, regexp.MustCompile(`^authelia_ac_`), token)

	assert.NoError(t, strategy.ValidateAuthorizeCode(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, token))
	assert.NoError(t, strategy.ValidateAuthorizeCode(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, token))
	assert.EqualError(t, strategy.ValidateAuthorizeCode(ctx, &fosite.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &fosite.DefaultSession{}}, token), "invalid_token")
	assert.NoError(t, strategy.ValidateAuthorizeCode(ctx, &fosite.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.AuthorizeCode: time.Now().Add(100 * time.Hour)}}}, token))
	assert.EqualError(t, strategy.ValidateAuthorizeCode(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.AuthorizeCode: time.Now().Add(-100 * time.Second)}}}, token), "invalid_token")

	token, signature, err = strategy.GenerateRefreshToken(ctx, &fosite.Request{})
	assert.NoError(t, err)

	assert.NotEmpty(t, token)
	assert.NotEmpty(t, signature)
	assert.Equal(t, signature, strategy.RefreshTokenSignature(ctx, token))
	assert.Regexp(t, regexp.MustCompile(`^authelia_rt_`), token)

	assert.NoError(t, strategy.ValidateRefreshToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, token))
	assert.NoError(t, strategy.ValidateRefreshToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, token))
	assert.NoError(t, strategy.ValidateRefreshToken(ctx, &fosite.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.RefreshToken: time.Now().Add(100 * time.Hour)}}}, token))
	assert.EqualError(t, strategy.ValidateRefreshToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.RefreshToken: time.Now().Add(-100 * time.Second)}}}, token), "invalid_token")

	token, signature, err = strategy.GenerateAccessToken(ctx, &fosite.Request{})
	assert.NoError(t, err)

	assert.NotEmpty(t, token)
	assert.NotEmpty(t, signature)
	assert.Equal(t, signature, strategy.AccessTokenSignature(ctx, token))
	assert.Regexp(t, regexp.MustCompile(`^authelia_at_`), token)

	assert.NoError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, token))
	assert.EqualError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &fosite.DefaultSession{}}, token), "invalid_token")
	assert.NoError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.AccessToken: time.Now().Add(100 * time.Hour)}}}, token))
	assert.EqualError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.AccessToken: time.Now().Add(-100 * time.Second)}}}, token), "invalid_token")

	token, signature, err = strategy.GenerateAccessToken(ctx, &fosite.Request{Client: &oidc.RegisteredClient{AccessTokenSignedResponseAlg: oidc.SigningAlgRSAUsingSHA256}})
	assert.Equal(t, "", token)
	assert.Equal(t, "", signature)
	assert.EqualError(t, err, "Session must be of type JWTSessionContainer but got type: <nil>")

	token, signature, err = strategy.GenerateAccessToken(ctx, &fosite.Request{Client: &oidc.RegisteredClient{AccessTokenSignedResponseAlg: oidc.SigningAlgRSAUsingSHA256}, Session: oidc.NewSession()})
	assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+$`), token)
	assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), signature)
	assert.True(t, strings.HasSuffix(token, signature))
	assert.NoError(t, err)

	assert.NoError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, token))
	assert.Equal(t, signature, strategy.AccessTokenSignature(ctx, token))
	assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, strings.Replace(token, signature, "qePeTyHu389VN_1woLEGR2v1LDJxUWhxrZZfDgUEf_hPtdnRKZv9fVLWJFNI06r87sC9Uu7IjuLqzAuqjwnE86BKZLYkMf780fPr-73Ohoq4jXUQI40uUodxaY4LVPuvq_5W2bAqLm5F03snKOYDQc_GQggek4SVmyDKqSUdvH4M5KXFhp2XyCu7BYv-retZG3K5Z0s_VS_tE8FF_S7_k1MXqSv_wwndmrn8ik-58bXlQe1bAHpvWCrtVQFJWEdtGaQoVDK40PHzLEaWEx47ys8jnAM4-rwNoBbxbP9NnK4Y1XRD1hzOpMYJ7UGa7hUwaIoOkmfEuhWGUZnNeyQRHQ", 1))), "Token signature mismatch. Check that you provided  a valid token in the right format. go-jose/go-jose: error in cryptographic primitive")

	token, signature, err = strategy.GenerateAccessToken(ctx, &fosite.Request{Client: &oidc.RegisteredClient{AccessTokenSignedResponseAlg: oidc.SigningAlgRSAUsingSHA256}, Session: &BadJWTSessionContainer{Session: &fosite.DefaultSession{}}})
	assert.EqualError(t, err, "JWT Claims must not be nil")
	assert.Empty(t, token)
	assert.Empty(t, signature)
}

type BadJWTSessionContainer struct {
	fosite.Session
}

func (c *BadJWTSessionContainer) GetJWTClaims() jwt.JWTClaimsContainer {
	return nil
}

func (c *BadJWTSessionContainer) GetJWTHeader() *jwt.Headers {
	return nil
}
