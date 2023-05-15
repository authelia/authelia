package oidc_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/token/hmac"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestHMACStrategy(t *testing.T) {
	goodsecret := []byte("R7VCSUfnKc7Y5zE84q6GstYqfMGjL4wM")
	secreta := []byte("a")

	config := &oidc.Config{
		TokenEntropy: 10,
		GlobalSecret: secreta,
		Lifespans: oidc.LifespanConfig{
			AccessToken:   time.Hour,
			RefreshToken:  time.Hour,
			AuthorizeCode: time.Minute,
		},
	}

	strategy := &oidc.HMACCoreStrategy{
		Enigma: &hmac.HMACStrategy{Config: config},
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
	assert.NoError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{}}, token))
	assert.EqualError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &fosite.DefaultSession{}}, token), "invalid_token")
	assert.NoError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.AccessToken: time.Now().Add(100 * time.Hour)}}}, token))
	assert.EqualError(t, strategy.ValidateAccessToken(ctx, &fosite.Request{RequestedAt: time.Now(), Session: &fosite.DefaultSession{ExpiresAt: map[fosite.TokenType]time.Time{fosite.AccessToken: time.Now().Add(-100 * time.Second)}}}, token), "invalid_token")
}
