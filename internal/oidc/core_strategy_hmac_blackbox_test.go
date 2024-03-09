package oidc_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/hmac"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestHMACCoreStrategy(t *testing.T) {
	goodsecret := []byte("R7VCSUfnKc7Y5zE84q6GstYqfMGjL4wM")
	secreta := []byte("a")

	config := &oidc.Config{
		TokenEntropy: 10,
		GlobalSecret: secreta,
		Lifespans: oidc.LifespansConfig{
			IdentityProvidersOpenIDConnectLifespanToken: schema.IdentityProvidersOpenIDConnectLifespanToken{
				AccessToken:   time.Hour,
				RefreshToken:  time.Hour,
				AuthorizeCode: time.Minute,
			},
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

	token, signature, err = strategy.GenerateAuthorizeCode(ctx, &oauthelia2.Request{})
	assert.EqualError(t, err, "secret for signing HMAC-SHA512/256 is expected to be 32 byte long, got 1 byte")
	assert.Empty(t, token)
	assert.Empty(t, signature)

	config.GlobalSecret = goodsecret

	token, signature, err = strategy.GenerateAuthorizeCode(ctx, &oauthelia2.Request{})
	assert.NoError(t, err)

	assert.NotEmpty(t, token)
	assert.NotEmpty(t, signature)
	assert.Equal(t, signature, strategy.AuthorizeCodeSignature(ctx, token))
	assert.Regexp(t, regexp.MustCompile(`^authelia_ac_`), token)

	assert.NoError(t, strategy.ValidateAuthorizeCode(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{}}, token))
	assert.NoError(t, strategy.ValidateAuthorizeCode(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{}}, token))
	assert.EqualError(t, strategy.ValidateAuthorizeCode(ctx, &oauthelia2.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &oauthelia2.DefaultSession{}}, token), "invalid_token")
	assert.NoError(t, strategy.ValidateAuthorizeCode(ctx, &oauthelia2.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &oauthelia2.DefaultSession{ExpiresAt: map[oauthelia2.TokenType]time.Time{oauthelia2.AuthorizeCode: time.Now().Add(100 * time.Hour)}}}, token))
	assert.EqualError(t, strategy.ValidateAuthorizeCode(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{ExpiresAt: map[oauthelia2.TokenType]time.Time{oauthelia2.AuthorizeCode: time.Now().Add(-100 * time.Second)}}}, token), "invalid_token")

	token, signature, err = strategy.GenerateRefreshToken(ctx, &oauthelia2.Request{})
	assert.NoError(t, err)

	assert.NotEmpty(t, token)
	assert.NotEmpty(t, signature)
	assert.Equal(t, signature, strategy.RefreshTokenSignature(ctx, token))
	assert.Regexp(t, regexp.MustCompile(`^authelia_rt_`), token)

	assert.NoError(t, strategy.ValidateRefreshToken(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{}}, token))
	assert.NoError(t, strategy.ValidateRefreshToken(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{}}, token))
	assert.NoError(t, strategy.ValidateRefreshToken(ctx, &oauthelia2.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &oauthelia2.DefaultSession{ExpiresAt: map[oauthelia2.TokenType]time.Time{oauthelia2.RefreshToken: time.Now().Add(100 * time.Hour)}}}, token))
	assert.EqualError(t, strategy.ValidateRefreshToken(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{ExpiresAt: map[oauthelia2.TokenType]time.Time{oauthelia2.RefreshToken: time.Now().Add(-100 * time.Second)}}}, token), "invalid_token")

	token, signature, err = strategy.GenerateAccessToken(ctx, &oauthelia2.Request{})
	assert.NoError(t, err)

	assert.NotEmpty(t, token)
	assert.NotEmpty(t, signature)
	assert.Equal(t, signature, strategy.AccessTokenSignature(ctx, token))
	assert.Regexp(t, regexp.MustCompile(`^authelia_at_`), token)

	assert.NoError(t, strategy.ValidateAccessToken(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{}}, token))
	assert.NoError(t, strategy.ValidateAccessToken(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{}}, token))
	assert.EqualError(t, strategy.ValidateAccessToken(ctx, &oauthelia2.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &oauthelia2.DefaultSession{}}, token), "invalid_token")
	assert.NoError(t, strategy.ValidateAccessToken(ctx, &oauthelia2.Request{RequestedAt: time.Now().Add(time.Hour * -2400), Session: &oauthelia2.DefaultSession{ExpiresAt: map[oauthelia2.TokenType]time.Time{oauthelia2.AccessToken: time.Now().Add(100 * time.Hour)}}}, token))
	assert.EqualError(t, strategy.ValidateAccessToken(ctx, &oauthelia2.Request{RequestedAt: time.Now(), Session: &oauthelia2.DefaultSession{ExpiresAt: map[oauthelia2.TokenType]time.Time{oauthelia2.AccessToken: time.Now().Add(-100 * time.Second)}}}, token), "invalid_token")

	badconfig := &BadGlobalSecretConfig{
		Config: config,
	}

	badstrategy := &oidc.HMACCoreStrategy{
		Enigma: &hmac.HMACStrategy{Config: badconfig},
		Config: badconfig,
	}

	token, signature, err = badstrategy.GenerateRefreshToken(ctx, &oauthelia2.Request{})
	assert.Equal(t, "", token)
	assert.Equal(t, "", signature)
	assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), "bad secret")

	token, signature, err = badstrategy.GenerateAccessToken(ctx, &oauthelia2.Request{})
	assert.Equal(t, "", token)
	assert.Equal(t, "", signature)
	assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), "bad secret")

	token, signature, err = badstrategy.GenerateAuthorizeCode(ctx, &oauthelia2.Request{})

	assert.Equal(t, "", token)
	assert.Equal(t, "", signature)
	assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), "bad secret")
}

type BadGlobalSecretConfig struct {
	*oidc.Config
}

func (*BadGlobalSecretConfig) GetGlobalSecret(ctx context.Context) ([]byte, error) {
	return nil, fmt.Errorf("bad secret")
}
