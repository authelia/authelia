package oidc

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/token/hmac"
	"github.com/stretchr/testify/assert"
)

func TestHMACStrategy(t *testing.T) {
	goodsecret := []byte("R7VCSUfnKc7Y5zE84q6GstYqfMGjL4wM")
	secreta := []byte("a")

	config := &Config{
		TokenEntropy: 10,
		GlobalSecret: secreta,
		Lifespans: LifespanConfig{
			AccessToken:   time.Hour,
			RefreshToken:  time.Hour,
			AuthorizeCode: time.Minute,
		},
	}

	strategy := &HMACCoreStrategy{
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

func TestHMACCoreStrategy_TrimPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		part     string
		expected string
	}{
		{"ShouldTrimAutheliaPrefix", "authelia_at_example", tokenPrefixPartAccessToken, "example"},
		{"ShouldTrimOryPrefix", "ory_at_example", tokenPrefixPartAccessToken, "example"},
		{"ShouldTrimOnlyAutheliaPrefix", "authelia_at_ory_at_example", tokenPrefixPartAccessToken, "ory_at_example"},
		{"ShouldTrimOnlyOryPrefix", "ory_at_authelia_at_example", tokenPrefixPartAccessToken, "authelia_at_example"},
		{"ShouldNotTrimGitHubPrefix", "gh_at_example", tokenPrefixPartAccessToken, "gh_at_example"},
	}

	strategy := &HMACCoreStrategy{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, strategy.trimPrefix(tc.have, tc.part))
		})
	}
}

func TestHMACCoreStrategy_GetSetPrefix(t *testing.T) {
	testCases := []struct {
		name        string
		have        string
		expectedSet string
		expectedGet string
	}{
		{"ShouldAddPrefix", "example", "authelia_%s_example", "authelia_%s_"},
	}

	strategy := &HMACCoreStrategy{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, part := range []string{tokenPrefixPartAccessToken, tokenPrefixPartAuthorizeCode, tokenPrefixPartRefreshToken} {
				t.Run(strings.ToUpper(part), func(t *testing.T) {
					assert.Equal(t, fmt.Sprintf(tc.expectedSet, part), strategy.setPrefix(tc.have, part))
					assert.Equal(t, fmt.Sprintf(tc.expectedGet, part), strategy.getPrefix(part))
				})
			}
		})
	}
}
