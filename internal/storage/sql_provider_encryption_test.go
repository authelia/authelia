package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
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

			provider := NewSQLiteProvider(config)
			require.NotNil(t, provider)

			_, err := provider.SchemaEncryptionCheckKey(context.Background(), false)

			assert.ErrorIs(t, err, ErrSchemaEncryptionVersionUnsupported)
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

	provider := NewSQLiteProvider(config)

	require.NotNil(t, provider)

	return provider
}
