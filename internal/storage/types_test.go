package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionValidationResult(t *testing.T) {
	result := &EncryptionValidationResult{
		InvalidCheckValue: false,
	}

	assert.True(t, result.Success())
	assert.True(t, result.Checked())

	result = &EncryptionValidationResult{
		InvalidCheckValue: true,
	}

	assert.False(t, result.Success())
	assert.True(t, result.Checked())

	result = &EncryptionValidationResult{
		InvalidCheckValue: false,
		Tables: map[string]EncryptionValidationTableResult{
			tableWebAuthnDevices: {
				Invalid: 10,
				Total:   20,
			},
		},
	}
	assert.Equal(t, "FAILURE", result.Tables[tableWebAuthnDevices].ResultDescriptor())

	assert.False(t, result.Success())
	assert.True(t, result.Checked())

	result = &EncryptionValidationResult{
		InvalidCheckValue: false,
		Tables: map[string]EncryptionValidationTableResult{
			tableWebAuthnDevices: {
				Error: fmt.Errorf("failed to check table"),
			},
		},
	}

	assert.False(t, result.Success())
	assert.False(t, result.Checked())
	assert.Equal(t, "N/A", result.Tables[tableWebAuthnDevices].ResultDescriptor())

	result = &EncryptionValidationResult{
		InvalidCheckValue: false,
		Tables: map[string]EncryptionValidationTableResult{
			tableWebAuthnDevices: {
				Total: 20,
			},
		},
	}

	assert.True(t, result.Success())
	assert.True(t, result.Checked())
	assert.Equal(t, "SUCCESS", result.Tables[tableWebAuthnDevices].ResultDescriptor())
}

func TestOAuth2SessionType(t *testing.T) {
	assert.Equal(t, "access token", OAuth2SessionTypeAccessToken.String())
	assert.Equal(t, tableOAuth2AccessTokenSession, OAuth2SessionTypeAccessToken.Table())

	assert.Equal(t, "authorization code", OAuth2SessionTypeAuthorizeCode.String())
	assert.Equal(t, tableOAuth2AuthorizeCodeSession, OAuth2SessionTypeAuthorizeCode.Table())

	assert.Equal(t, "openid connect", OAuth2SessionTypeOpenIDConnect.String())
	assert.Equal(t, tableOAuth2OpenIDConnectSession, OAuth2SessionTypeOpenIDConnect.Table())

	assert.Equal(t, "pushed authorization request context", OAuth2SessionTypePAR.String())
	assert.Equal(t, tableOAuth2PARContext, OAuth2SessionTypePAR.Table())

	assert.Equal(t, "pkce challenge", OAuth2SessionTypePKCEChallenge.String())
	assert.Equal(t, tableOAuth2PKCERequestSession, OAuth2SessionTypePKCEChallenge.Table())

	assert.Equal(t, "refresh token", OAuth2SessionTypeRefreshToken.String())
	assert.Equal(t, tableOAuth2RefreshTokenSession, OAuth2SessionTypeRefreshToken.Table())

	assert.Equal(t, "invalid", OAuth2SessionType(-1).String())
	assert.Equal(t, "", OAuth2SessionType(-1).Table())
}
