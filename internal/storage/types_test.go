package storage

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

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
			tableWebAuthnCredentials: {
				Invalid: 10,
				Total:   20,
			},
		},
	}
	assert.Equal(t, "FAILURE", result.Tables[tableWebAuthnCredentials].ResultDescriptor())

	assert.False(t, result.Success())
	assert.True(t, result.Checked())

	result = &EncryptionValidationResult{
		InvalidCheckValue: false,
		Tables: map[string]EncryptionValidationTableResult{
			tableWebAuthnCredentials: {
				Error: fmt.Errorf("failed to check table"),
			},
		},
	}

	assert.False(t, result.Success())
	assert.False(t, result.Checked())
	assert.Equal(t, "N/A", result.Tables[tableWebAuthnCredentials].ResultDescriptor())

	result = &EncryptionValidationResult{
		InvalidCheckValue: false,
		Tables: map[string]EncryptionValidationTableResult{
			tableWebAuthnCredentials: {
				Total: 20,
			},
		},
	}

	assert.True(t, result.Success())
	assert.True(t, result.Checked())
	assert.Equal(t, "SUCCESS", result.Tables[tableWebAuthnCredentials].ResultDescriptor())
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

	assert.Equal(t, "device code", OAuth2SessionTypeDeviceAuthorizeCode.String())
	assert.Equal(t, tableOAuth2DeviceCodeSession, OAuth2SessionTypeDeviceAuthorizeCode.Table())

	assert.Equal(t, "invalid", OAuth2SessionType(-1).String())
	assert.Equal(t, "", OAuth2SessionType(-1).Table())
}

func TestBanExpiresExpiredExpiration(t *testing.T) {
	expiredAt := time.Date(2024, 7, 4, 12, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2024, 8, 4, 12, 0, 0, 0, time.UTC)
	zeroTime := time.Unix(0, 0)

	testCases := []struct {
		name     string
		have     banExpiresExpired
		expected time.Time
	}{
		{
			name: "ShouldReturnExpiredWhenRevokedAndExpiredValid",
			have: banExpiresExpired{
				Revoked: true,
				Expired: sql.NullTime{Time: expiredAt, Valid: true},
				Expires: sql.NullTime{Time: expiresAt, Valid: true},
			},
			expected: expiredAt,
		},
		{
			name: "ShouldReturnExpiresWhenRevokedButExpiredInvalid",
			have: banExpiresExpired{
				Revoked: true,
				Expired: sql.NullTime{Valid: false},
				Expires: sql.NullTime{Time: expiresAt, Valid: true},
			},
			expected: expiresAt,
		},
		{
			name: "ShouldReturnExpiresWhenNotRevokedAndExpiresValid",
			have: banExpiresExpired{
				Revoked: false,
				Expired: sql.NullTime{Time: expiredAt, Valid: true},
				Expires: sql.NullTime{Time: expiresAt, Valid: true},
			},
			expected: expiresAt,
		},
		{
			name: "ShouldReturnZeroWhenNoTimesValid",
			have: banExpiresExpired{
				Revoked: false,
				Expired: sql.NullTime{Valid: false},
				Expires: sql.NullTime{Valid: false},
			},
			expected: zeroTime,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.Expiration())
		})
	}
}
