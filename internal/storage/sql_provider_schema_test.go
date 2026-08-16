package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldReturnErrOnTargetSameAsCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, 2, 2),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 2, 2))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))
}

func TestShouldReturnErrOnUpMigrationTargetVersionLessThanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, 0, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 0, LatestVersion))

	assert.NoError(t,
		schemaMigrateChecks(providerPostgres, true, LatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, 0, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 0, LatestVersion))

	assert.NoError(t,
		schemaMigrateChecks(providerSQLite, true, LatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, 0, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 0, LatestVersion))

	assert.NoError(t,
		schemaMigrateChecks(providerMySQL, true, LatestVersion, 0))
}

func TestMigrationUpShouldReturnErrOnAlreadyLatest(t *testing.T) {
	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest, LatestVersion))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest, LatestVersion))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest, LatestVersion))
}

func TestShouldReturnErrOnVersionDoesntExits(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest-1, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, LatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest-1, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, LatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest-1, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, LatestVersion))
}

func TestMigrationDownShouldReturnErrOnTargetLessThan1(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, -4, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -4))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, -2, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, -2, LatestVersion),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, -1, LatestVersion),
		"schema migration down to pre1 is no longer supported: you must use an older version of authelia to perform this migration: you should downgrade to schema version 1 using the current authelia version then use the suggested authelia version to downgrade to pre1: the suggested authelia version is 4.37.2")
}

func TestMigrationDownShouldReturnErrOnCurrentLessThan0(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, LatestVersion, -1),
		"schema migration up from pre1 is no longer supported: you must use an older version of authelia to perform this migration: the suggested authelia version is 4.37.2")
}

func TestMigrationDownShouldReturnErrOnTargetVersionGreaterThanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, LatestVersion, 0),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, LatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, LatestVersion, 0),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, LatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, LatestVersion, 0),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, LatestVersion, 0))
}

func TestShouldReturnErrWhenCurrentIsGreaterThanLatest(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, LatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, LatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, LatestVersion))
}

func TestSchemaVersionToString(t *testing.T) {
	assert.Equal(t, "unknown", SchemaVersionToString(-2))
	assert.Equal(t, "pre1", SchemaVersionToString(-1))
	assert.Equal(t, "N/A", SchemaVersionToString(0))
	assert.Equal(t, "1", SchemaVersionToString(1))
	assert.Equal(t, "2", SchemaVersionToString(2))
}

func TestSchemaMigrateFromLegacyWithWebAuthnCredential(t *testing.T) {
	provider := newTestSQLiteProvider(t)

	ctx := context.Background()

	require.NoError(t, provider.SchemaMigrate(ctx, true, 23))

	legacy := utils.DeriveLegacyCryptographicKey([]byte("authelia-test-key-not-a-secret-authelia-test-key-not-a-secret"))

	checkValue, err := utils.Encrypt([]byte("test-check-value"), nil, legacy)
	require.NoError(t, err)

	_, err = provider.db.ExecContext(ctx, provider.sqlUpsertEncryptionValue, encryptionNameCheck, checkValue)
	require.NoError(t, err)

	publicKey, err := utils.Encrypt([]byte("fake-public-key"), nil, legacy)
	require.NoError(t, err)

	_, err = provider.db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO %s (created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, tableWebAuthnCredentials),
		time.Now(), time.Now(), "example.com", "john", "test", "a2lk", nil, "packed", "cross-platform", "", 0, false, false, false, false, false, false, false, publicKey)
	require.NoError(t, err)

	require.NoError(t, provider.SchemaMigrate(ctx, true, SchemaLatest))

	credentials, err := provider.LoadWebAuthnCredentialsByUsername(ctx, "example.com", "john")

	require.NoError(t, err)
	require.Len(t, credentials, 1)
	assert.Equal(t, []byte("fake-public-key"), credentials[0].PublicKey)
}

func TestSchemaMigrateToRowScopedAAD(t *testing.T) {
	testCases := []struct {
		name  string
		prior int
	}{
		{
			name:  "ShouldMigrateFromLegacy",
			prior: 23,
		},
		{
			name:  "ShouldMigrateFromColumnScoped",
			prior: 25,
		},
	}

	totpSecret := []byte("JBSWY3DPEHPK3PXP")
	sessionData := []byte(`{"access":"token"}`)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProvider(t)

			ctx := context.Background()

			require.NoError(t, provider.SchemaMigrate(ctx, true, tc.prior))

			key, aad := provider.keys.encryption, provider.aad

			if tc.prior < schemaVersionEncryptionKeyDerivation {
				provider.keys.encryption = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))
				provider.aad = aadNone
			} else {
				provider.aad = aadColumn
			}

			require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
				CreatedAt: time.Now().Truncate(time.Second),
				Username:  "john",
				Issuer:    "Authelia",
				Algorithm: "SHA1",
				Digits:    6,
				Period:    30,
				Secret:    totpSecret,
			}))

			require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAccessToken, model.OAuth2Session{
				RequestID: "req-123",
				Signature: "sig-123",
				Session:   sessionData,
			}))

			provider.keys.encryption, provider.aad = key, aad

			require.NoError(t, provider.SchemaMigrate(ctx, true, SchemaLatest))

			result, err := provider.SchemaEncryptionCheckKey(ctx, true)

			require.NoError(t, err)
			assert.False(t, result.InvalidCheckValue)

			for table, tableResult := range result.Tables {
				assert.NoError(t, tableResult.Error, table)
				assert.Equal(t, 0, tableResult.Invalid, table)
			}

			totp, err := provider.LoadTOTPConfiguration(ctx, "john")

			require.NoError(t, err)
			assert.Equal(t, totpSecret, totp.Secret)

			session, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeAccessToken, "sig-123")

			require.NoError(t, err)
			assert.Equal(t, sessionData, session.Session)
		})
	}
}

func TestSchemaMigrateDownFromRowScopedAAD(t *testing.T) {
	testCases := []struct {
		name   string
		target int
	}{
		{
			name:   "ShouldMigrateDownToColumnScoped",
			target: 25,
		},
		{
			name:   "ShouldMigrateDownToLegacy",
			target: 24,
		},
	}

	totpSecret := []byte("JBSWY3DPEHPK3PXP")
	sessionData := []byte(`{"access":"token"}`)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProvider(t)

			ctx := context.Background()

			require.NoError(t, provider.SchemaMigrate(ctx, true, SchemaLatest))

			require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
				CreatedAt: time.Now().Truncate(time.Second),
				Username:  "john",
				Issuer:    "Authelia",
				Algorithm: "SHA1",
				Digits:    6,
				Period:    30,
				Secret:    totpSecret,
			}))

			require.NoError(t, provider.SaveOAuth2Session(ctx, OAuth2SessionTypeAccessToken, model.OAuth2Session{
				RequestID: "req-123",
				Signature: "sig-123",
				Session:   sessionData,
			}))

			require.NoError(t, provider.SchemaMigrate(ctx, false, tc.target))

			result, err := provider.SchemaEncryptionCheckKey(ctx, true)

			require.NoError(t, err)
			assert.False(t, result.InvalidCheckValue)

			for table, tableResult := range result.Tables {
				assert.NoError(t, tableResult.Error, table)
				assert.Equal(t, 0, tableResult.Invalid, table)
			}
		})
	}
}
