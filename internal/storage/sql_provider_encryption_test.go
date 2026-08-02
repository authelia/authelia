package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestSchemaEncryptionCheckKey(t *testing.T) {
	testCases := []struct {
		name    string
		verbose bool
	}{
		{"ShouldSucceedNonVerbose", false},
		{"ShouldSucceedVerbose", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			result, err := provider.SchemaEncryptionCheckKey(ctx, tc.verbose)

			assert.NoError(t, err)
			assert.True(t, result.Success())
		})
	}
}

func TestSchemaEncryptionCheckKeyWithData(t *testing.T) {
	testCases := []struct {
		name     string
		seedTOTP bool
		verbose  bool
	}{
		{"ShouldSucceedEmptyVerbose", false, true},
		{"ShouldSucceedWithTOTPVerbose", true, true},
		{"ShouldSucceedWithTOTPNonVerbose", true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			if tc.seedTOTP {
				require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
					CreatedAt: time.Now().Truncate(time.Second),
					Username:  "john",
					Issuer:    "Authelia",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    []byte("JBSWY3DPEHPK3PXP"),
				}))
			}

			result, err := provider.SchemaEncryptionCheckKey(ctx, tc.verbose)

			assert.NoError(t, err)
			assert.True(t, result.Success())

			if tc.verbose {
				assert.NotEmpty(t, result.Tables)
			}
		})
	}
}

func TestSchemaEncryptionChangeKey(t *testing.T) {
	testCases := []struct {
		name   string
		newKey string
		err    string
	}{
		{
			"ShouldSucceedChangeKey",
			"authelia-new-test-key-not-a-secret-authelia-new-key",
			"",
		},
		{
			"ShouldErrSameKey",
			"authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			"error changing the storage encryption key: the old key and the new key are the same",
		},
		{
			"ShouldErrEmptyKey",
			"",
			"error deriving cryptographic key: value is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			err := provider.SchemaEncryptionChangeKey(ctx, tc.newKey)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestSchemaEncryptionChangeKeyWithData(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"ShouldSucceedChangeKeyWithTOTPData"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
				CreatedAt: time.Now().Truncate(time.Second),
				Username:  "john",
				Issuer:    "Authelia",
				Algorithm: "SHA1",
				Digits:    6,
				Period:    30,
				Secret:    []byte("JBSWY3DPEHPK3PXP"),
			}))

			err := provider.SchemaEncryptionChangeKey(ctx, "authelia-new-test-key-not-a-secret-authelia-new-key")

			assert.NoError(t, err)
		})
	}
}

func TestSchemaEncryptionChangeKeyShouldRollbackOnCorruptData(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"ShouldRollbackWhenTOTPSecretCannotBeDecrypted"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
				CreatedAt: time.Now().Truncate(time.Second),
				Username:  "john",
				Issuer:    "Authelia",
				Algorithm: "SHA1",
				Digits:    6,
				Period:    30,
				Secret:    []byte("JBSWY3DPEHPK3PXP"),
			}))

			_, err := provider.db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = ?", tableTOTPConfigurations, columnSecret), []byte("this-is-not-valid-ciphertext"))
			require.NoError(t, err)

			err = provider.SchemaEncryptionChangeKey(ctx, "authelia-new-test-key-not-a-secret-authelia-new-key")
			assert.EqualError(t, err, "error changing the storage encryption key: error decrypting TOTP configuration secret with id '1': cipher: message authentication failed")

			var secret []byte

			require.NoError(t, provider.db.GetContext(ctx, &secret, fmt.Sprintf("SELECT %s FROM %s WHERE username = ?", columnSecret, tableTOTPConfigurations), "john"))
			assert.Equal(t, []byte("this-is-not-valid-ciphertext"), secret)
		})
	}
}

func TestSchemaEncryptionChangeKeyShouldSkipEmptyCachedData(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"ShouldSkipCachedDataRowWithEmptyValue"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			_, err := provider.db.ExecContext(ctx, provider.sqlUpsertCachedData, "empty-cache", time.Now(), true, []byte{})
			require.NoError(t, err)

			require.NoError(t, provider.SchemaEncryptionChangeKey(ctx, "authelia-new-test-key-not-a-secret-authelia-new-key"))
		})
	}
}

func TestSchemaEncryptionCheckKeyVersionUnsupported(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"ShouldErrWhenSchemaNotMigrated"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.Configuration{
				Storage: schema.Storage{
					EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
					Local: &schema.StorageLocal{
						Path: filepath.Join(t.TempDir(), "db.sqlite3"),
					},
				},
			}

			provider, err := NewSQLiteProvider(config)

			require.NoError(t, err)
			require.NotNil(t, provider)

			_, err = provider.SchemaEncryptionCheckKey(context.Background(), false)

			assert.ErrorIs(t, err, ErrSchemaEncryptionVersionUnsupported)
		})
	}
}

func TestSchemaEncryptionUpgradeFromLegacyKey(t *testing.T) {
	testCases := []struct {
		name     string
		seedTOTP bool
	}{
		{"ShouldUpgradeCheckValueOnly", false},
		{"ShouldUpgradeCheckValueAndData", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)

			legacyKey := utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))

			ctx := context.Background()

			require.NoError(t, provider.SchemaMigrate(ctx, true, schemaVersionEncryptionKeyDerivation-1))

			checkValue, err := utils.Encrypt([]byte(uuid.Must(uuid.NewRandom()).String()), nil, legacyKey)
			require.NoError(t, err)

			_, err = provider.db.ExecContext(ctx, provider.sqlUpsertEncryptionValue, encryptionNameCheck, checkValue)
			require.NoError(t, err)

			if tc.seedTOTP {
				secret, err := utils.Encrypt([]byte("JBSWY3DPEHPK3PXP"), nil, legacyKey)
				require.NoError(t, err)

				_, err = provider.db.ExecContext(ctx, provider.sqlUpsertTOTPConfig,
					time.Now().Truncate(time.Second), sql.NullTime{},
					"john", "Authelia",
					"SHA1", 6, 30, secret)
				require.NoError(t, err)
			}

			require.NoError(t, provider.StartupCheck())

			version, err := provider.SchemaVersion(ctx)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, version, schemaVersionEncryptionKeyDerivation)

			result, err := provider.SchemaEncryptionCheckKey(ctx, true)
			require.NoError(t, err)
			assert.True(t, result.Success())

			if tc.seedTOTP {
				config, err := provider.LoadTOTPConfiguration(ctx, "john")
				require.NoError(t, err)
				assert.Equal(t, []byte("JBSWY3DPEHPK3PXP"), config.Secret)
			}
		})
	}
}

func TestStorageUserTOTPShouldRoundTripWithoutStartupCheck(t *testing.T) {
	config := &schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	}

	ctx := context.Background()

	migrator, err := NewSQLiteProvider(config)
	require.NoError(t, err)
	require.NoError(t, migrator.SchemaMigrate(ctx, true, SchemaLatest))
	require.NoError(t, migrator.Close())

	provider, err := NewSQLiteProvider(config)
	require.NoError(t, err)

	require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
		CreatedAt: time.Now().Truncate(time.Second),
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    []byte("JBSWY3DPEHPK3PXP"),
	}))

	loaded, err := provider.LoadTOTPConfiguration(ctx, "john")
	require.NoError(t, err)
	assert.Equal(t, []byte("JBSWY3DPEHPK3PXP"), loaded.Secret)

	configs, err := provider.LoadTOTPConfigurations(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, []byte("JBSWY3DPEHPK3PXP"), configs[0].Secret)
}

func TestSchemaEncryptionChangeKeyWithAllData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite3")

	const (
		oldKey = "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret"
		newKey = "authelia-new-test-key-not-a-secret-authelia-new-key-value-ok"
	)

	newProvider := func(key string) *SQLiteProvider {
		config := &schema.Configuration{
			Storage: schema.Storage{
				EncryptionKey: key,
				Local:         &schema.StorageLocal{Path: path},
			},
		}

		provider, err := NewSQLiteProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		return provider
	}

	ctx := context.Background()

	provider := newProvider(oldKey)
	require.NoError(t, provider.StartupCheck())

	require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
		CreatedAt: time.Now().Truncate(time.Second),
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    []byte("JBSWY3DPEHPK3PXP"),
	}))

	require.NoError(t, provider.SaveWebAuthnCredential(ctx, model.WebAuthnCredential{
		CreatedAt:       time.Now().Truncate(time.Second),
		RPID:            "example.com",
		Username:        "john",
		Description:     "my-key",
		KID:             model.NewBase64([]byte("kid-1")),
		AttestationType: "none",
		Attachment:      "cross-platform",
		PublicKey:       []byte("fake-public-key"),
		Attestation:     []byte("fake-attestation"),
	}))

	_, err := provider.SaveOneTimeCode(ctx, model.OneTimeCode{
		PublicID:  uuid.Must(uuid.NewRandom()),
		IssuedAt:  time.Now().Truncate(time.Second),
		IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
		ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
		Username:  "john",
		Intent:    "reset_password",
		Code:      []byte("123456"),
	})
	require.NoError(t, err)

	require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAccessToken, model.OAuth2Session{
		ChallengeID:     model.MustNullUUID(model.NewRandomNullUUID()),
		RequestID:       "req-123",
		ClientID:        "test-client",
		Signature:       "sig-123",
		Subject:         sql.NullString{Valid: true, String: "john"},
		Active:          true,
		RequestedScopes: model.StringSlicePipeDelimited{"openid"},
		GrantedScopes:   model.StringSlicePipeDelimited{"openid"},
		Session:         []byte(`{"access":"token"}`),
	}))

	require.NoError(t, provider.SaveOAuth2DeviceCodeSession(ctx, &model.OAuth2DeviceCodeSession{
		Signature:         "dev-sig-123",
		RequestID:         "dev-req-123",
		ClientID:          "test-client",
		UserCodeSignature: "user-code-123",
		Active:            true,
		RequestedScopes:   model.StringSlicePipeDelimited{"openid"},
		GrantedScopes:     model.StringSlicePipeDelimited{"openid"},
		Session:           []byte(`{"device":"code"}`),
		RequestedAt:       time.Now().Truncate(time.Second),
	}))

	require.NoError(t, provider.SaveOAuth2PushedAuthorizationSession(ctx, model.OAuth2PushedAuthorizationSession{
		Signature:   "par-sig-123",
		RequestID:   "par-req-123",
		ClientID:    "test-client",
		RequestedAt: time.Now().Truncate(time.Second),
		Session:     []byte(`{"par":"session"}`),
	}))

	require.NoError(t, provider.SaveCachedData(ctx, model.CachedData{
		Name:      "cache-key",
		Value:     []byte("cache-value"),
		Encrypted: true,
	}))

	require.NoError(t, provider.SchemaEncryptionChangeKey(ctx, newKey))
	require.NoError(t, provider.Close())

	provider = newProvider(newKey)
	require.NoError(t, provider.StartupCheck())

	result, err := provider.SchemaEncryptionCheckKey(ctx, true)
	require.NoError(t, err)
	assert.True(t, result.Success())

	totp, err := provider.LoadTOTPConfiguration(ctx, "john")
	require.NoError(t, err)
	assert.Equal(t, []byte("JBSWY3DPEHPK3PXP"), totp.Secret)

	credentials, err := provider.LoadWebAuthnCredentialsByUsername(ctx, "example.com", "john")
	require.NoError(t, err)
	require.Len(t, credentials, 1)
	assert.Equal(t, []byte("fake-public-key"), credentials[0].PublicKey)
	assert.Equal(t, []byte("fake-attestation"), credentials[0].Attestation)

	code, err := provider.LoadOneTimeCode(ctx, "john", model.NewIP(net.ParseIP("127.0.0.1")), "reset_password", "123456")
	require.NoError(t, err)
	assert.Equal(t, []byte("123456"), code.Code)

	session, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeAccessToken, "sig-123")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"access":"token"}`), session.Session)

	device, err := provider.LoadOAuth2DeviceCodeSession(ctx, "dev-sig-123")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"device":"code"}`), device.Session)

	par, err := provider.LoadOAuth2PushedAuthorizationSession(ctx, "par-sig-123")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"par":"session"}`), par.Session)

	cached, err := provider.LoadCachedData(ctx, "cache-key")
	require.NoError(t, err)
	assert.Equal(t, []byte("cache-value"), cached.Value)

	require.NoError(t, provider.Close())
}

func TestSchemaEncryptionCheckKeyWithInvalidData(t *testing.T) {
	testCases := []struct {
		name    string
		table   string
		column  string
		corrupt func(t *testing.T, provider *SQLiteProvider, ctx context.Context)
	}{
		{
			name:   "ShouldReportInvalidTOTPSecret",
			table:  tableTOTPConfigurations,
			column: columnSecret,
			corrupt: func(t *testing.T, provider *SQLiteProvider, ctx context.Context) {
				require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
					CreatedAt: time.Now().Truncate(time.Second),
					Username:  "john",
					Issuer:    "Authelia",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    []byte("JBSWY3DPEHPK3PXP"),
				}))
			},
		},
		{
			name:   "ShouldReportInvalidCachedDataValue",
			table:  tableCachedData,
			column: columnValue,
			corrupt: func(t *testing.T, provider *SQLiteProvider, ctx context.Context) {
				require.NoError(t, provider.SaveCachedData(ctx, model.CachedData{
					Name:      "cache-key",
					Value:     []byte("cache-value"),
					Encrypted: true,
				}))
			},
		},
		{
			name:   "ShouldReportInvalidOneTimeCode",
			table:  tableOneTimeCode,
			column: columnCode,
			corrupt: func(t *testing.T, provider *SQLiteProvider, ctx context.Context) {
				_, err := provider.SaveOneTimeCode(ctx, model.OneTimeCode{
					PublicID:  uuid.Must(uuid.NewRandom()),
					IssuedAt:  time.Now().Truncate(time.Second),
					IssuedIP:  model.NewIP(net.ParseIP("127.0.0.1")),
					ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
					Username:  "john",
					Intent:    "reset_password",
					Code:      []byte("123456"),
				})
				require.NoError(t, err)
			},
		},
		{
			name:   "ShouldReportInvalidWebAuthnPublicKey",
			table:  tableWebAuthnCredentials,
			column: "public_key",
			corrupt: func(t *testing.T, provider *SQLiteProvider, ctx context.Context) {
				require.NoError(t, provider.SaveWebAuthnCredential(ctx, model.WebAuthnCredential{
					CreatedAt:       time.Now().Truncate(time.Second),
					RPID:            "example.com",
					Username:        "john",
					Description:     "my-key",
					KID:             model.NewBase64([]byte("kid-1")),
					AttestationType: "none",
					Attachment:      "cross-platform",
					PublicKey:       []byte("fake-public-key"),
					Attestation:     []byte("fake-attestation"),
				}))
			},
		},
		{
			name:   "ShouldReportInvalidOAuth2SessionData",
			table:  tableOAuth2AccessTokenSession,
			column: columnSessionData,
			corrupt: func(t *testing.T, provider *SQLiteProvider, ctx context.Context) {
				require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAccessToken, model.OAuth2Session{
					ChallengeID:     model.MustNullUUID(model.NewRandomNullUUID()),
					RequestID:       "req-123",
					ClientID:        "test-client",
					Signature:       "sig-123",
					Subject:         sql.NullString{Valid: true, String: "john"},
					Active:          true,
					RequestedScopes: model.StringSlicePipeDelimited{"openid"},
					GrantedScopes:   model.StringSlicePipeDelimited{"openid"},
					Session:         []byte(`{"access":"token"}`),
				}))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			tc.corrupt(t, provider, ctx)

			_, err := provider.db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = ?", tc.table, tc.column), []byte("this-is-not-valid-ciphertext"))
			require.NoError(t, err)

			result, err := provider.SchemaEncryptionCheckKey(ctx, true)
			require.NoError(t, err)

			assert.False(t, result.Success())
			assert.NotZero(t, result.Tables[tc.table].Invalid)
		})
	}
}

func TestSchemaEncryptionCheckKeyShouldReportTableQueryErrors(t *testing.T) {
	testCases := []struct {
		name  string
		table string
	}{
		{"ShouldReportOneTimeCodeQueryError", tableOneTimeCode},
		{"ShouldReportTOTPQueryError", tableTOTPConfigurations},
		{"ShouldReportWebAuthnQueryError", tableWebAuthnCredentials},
		{"ShouldReportCachedDataQueryError", tableCachedData},
		{"ShouldReportOAuth2SessionQueryError", tableOAuth2AccessTokenSession},
		{"ShouldReportEncryptionQueryError", tableEncryption},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProviderWithEncryption(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			_, err := provider.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE %s", tc.table))
			require.NoError(t, err)

			result, err := provider.SchemaEncryptionCheckKey(ctx, true)
			require.NoError(t, err)

			assert.False(t, result.Success())
			assert.Error(t, result.Tables[tc.table].Error)
		})
	}
}

func newTestSQLiteProviderWithEncryption(t *testing.T) *SQLiteProvider {
	t.Helper()

	config := &schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	}

	provider, err := NewSQLiteProvider(config)

	require.NoError(t, err)
	require.NotNil(t, provider)

	return provider
}
