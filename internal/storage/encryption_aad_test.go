package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionAAD(t *testing.T) {
	testCases := []struct {
		name     string
		aad      EncryptionAAD
		expected string
	}{
		{
			name:     "ShouldReturnNothingForNone",
			aad:      aadNone,
			expected: "",
		},
		{
			name:     "ShouldReturnColumnScopeForColumn",
			aad:      aadColumn,
			expected: "authelia:storage:one_time_code:code",
		},
		{
			name:     "ShouldReturnRowScopeForRow",
			aad:      aadRow,
			expected: "authelia:storage:one_time_code:code:abc123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.aad.Get(tableOneTimeCode, columnCode, "abc123")

			if tc.expected == "" {
				assert.Nil(t, actual)
			} else {
				assert.Equal(t, tc.expected, string(actual))
			}
		})
	}
}

func TestEncryptionAADIsEncryptionAAD(t *testing.T) {
	testCases := []struct {
		name string
		aad  EncryptionAAD
	}{
		{
			name: "ShouldImplementNone",
			aad:  aadNone,
		},
		{
			name: "ShouldImplementColumn",
			aad:  aadColumn,
		},
		{
			name: "ShouldImplementRow",
			aad:  aadRow,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tc.aad.IsEncryptionAAD())
		})
	}
}

func TestAADForSchemaVersion(t *testing.T) {
	testCases := []struct {
		name     string
		version  int
		expected EncryptionAAD
	}{
		{
			name:     "ShouldReturnNoneForFreshSchema",
			version:  0,
			expected: aadNone,
		},
		{
			name:     "ShouldReturnNoneForSchemaOne",
			version:  1,
			expected: aadNone,
		},
		{
			name:     "ShouldReturnNoneForSchemaBelowKeyDerivation",
			version:  schemaVersionEncryptionKeyDerivation - 1,
			expected: aadNone,
		},
		{
			name:     "ShouldReturnColumnForSchemaKeyDerivation",
			version:  schemaVersionEncryptionKeyDerivation,
			expected: aadColumn,
		},
		{
			name:     "ShouldReturnRowForSchemaRowScoped",
			version:  schemaVersionEncryptionAADRowScoped,
			expected: aadRow,
		},
		{
			name:     "ShouldReturnRowForSchemaAboveRowScoped",
			version:  schemaVersionEncryptionAADRowScoped + 1,
			expected: aadRow,
		},
		{
			name:     "ShouldReturnRowForLatestSchema",
			version:  SchemaLatest,
			expected: aadRow,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, aadForSchemaVersion(tc.version))
		})
	}
}

func TestEncryptionAADIssuer(t *testing.T) {
	testCases := []struct {
		name     string
		aad      EncryptionAAD
		expected string
	}{
		{
			name:     "ShouldReturnNothingForNone",
			aad:      aadNone,
			expected: "",
		},
		{
			name:     "ShouldReturnLegacyInterleavedOrderForColumn",
			aad:      aadColumn,
			expected: "authelia:storage:webauthn_credentials:example.com:public_key",
		},
		{
			name:     "ShouldReturnRowThenIssuerForRow",
			aad:      aadRow,
			expected: "authelia:storage:webauthn_credentials:public_key:a2lk:example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.aad.GetIssuer(tableWebAuthnCredentials, "public_key", "a2lk", "example.com")

			if tc.expected == "" {
				assert.Nil(t, actual)
			} else {
				assert.Equal(t, tc.expected, string(actual))
			}
		})
	}
}

func TestEncryptionAADColumnIgnoresRow(t *testing.T) {
	testCases := []struct {
		name string
		row  string
	}{
		{
			name: "ShouldIgnoreEmptyRow",
			row:  "",
		},
		{
			name: "ShouldIgnorePopulatedRow",
			row:  "abc123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, "authelia:storage:encryption:value", string(aadColumn.Get(tableEncryption, columnValue, tc.row)))
		})
	}
}

func TestEncryptionAADRowMatchesLegacyForColumnPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		table    string
		column   string
		row      string
		expected string
	}{
		{
			name:     "ShouldScopeTOTPByUsername",
			table:    tableTOTPConfigurations,
			column:   columnSecret,
			row:      "john",
			expected: "authelia:storage:totp_configurations:secret:john",
		},
		{
			name:     "ShouldScopeEncryptionByName",
			table:    tableEncryption,
			column:   columnValue,
			row:      "hmac_key_otc",
			expected: "authelia:storage:encryption:value:hmac_key_otc",
		},
		{
			name:     "ShouldScopeCachedDataByName",
			table:    tableCachedData,
			column:   columnValue,
			row:      "example",
			expected: "authelia:storage:cached_data:value:example",
		},
		{
			name:     "ShouldScopePARBySignatureUsingAADTableName",
			table:    tableAADPushedAuthorizationRequestSession,
			column:   columnSessionData,
			row:      "sig",
			expected: "authelia:storage:oauth2_pushed_authorization_session:session_data:sig",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(aadRow.Get(tc.table, tc.column, tc.row)))
		})
	}
}
