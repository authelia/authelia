package storage

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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

			saveLegacyOAuth2AccessTokenSession(t, ctx, provider, model.OAuth2Session{
				RequestID: "req-123",
				Signature: "sig-123",
				Session:   sessionData,
			})

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

func TestSchemaMigrateUpToColumnScopedFromLegacy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite3")

	totpSecret := []byte("JBSWY3DPEHPK3PXP")
	sessionData := []byte(`{"access":"token"}`)

	ctx := context.Background()

	provider := newTestSQLiteProviderAtPath(t, path)

	require.NoError(t, provider.SchemaMigrate(ctx, true, schemaVersionEncryptionKeyDerivation-1))

	key, aad := provider.keys.encryption, provider.aad

	provider.keys.encryption = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))
	provider.aad = aadNone

	require.NoError(t, provider.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
		CreatedAt: time.Now().Truncate(time.Second),
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    totpSecret,
	}))

	saveLegacyOAuth2AccessTokenSession(t, ctx, provider, model.OAuth2Session{
		RequestID: "req-123",
		Signature: "sig-123",
		Session:   sessionData,
	})

	provider.keys.encryption, provider.aad = key, aad

	require.NoError(t, provider.Close())

	provider = newTestSQLiteProviderAtPath(t, path)

	require.NoError(t, provider.SchemaMigrate(ctx, true, schemaVersionEncryptionKeyDerivation))

	version, err := provider.SchemaVersion(ctx)

	require.NoError(t, err)
	assert.Equal(t, schemaVersionEncryptionKeyDerivation, version)

	result, err := provider.SchemaEncryptionCheckKey(ctx, true)

	require.NoError(t, err)
	assert.False(t, result.InvalidCheckValue)

	for table, tableResult := range result.Tables {
		assert.NoError(t, tableResult.Error, table)
		assert.Equal(t, 0, tableResult.Invalid, table)
	}

	provider.aad = aadColumn

	totp, err := provider.LoadTOTPConfiguration(ctx, "john")

	require.NoError(t, err)
	assert.Equal(t, totpSecret, totp.Secret)

	decrypted := loadLegacyOAuth2AccessTokenSessionData(t, ctx, provider, "sig-123")

	assert.Equal(t, sessionData, decrypted)

	require.NoError(t, provider.Close())
}

func TestSchemaMigrateDownThroughColumnScopedThenBackUp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite3")

	totpSecret := []byte("JBSWY3DPEHPK3PXP")
	sessionData := []byte(`{"access":"token"}`)

	ctx := context.Background()

	provider := newTestSQLiteProviderAtPath(t, path)

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

	require.NoError(t, provider.Close())

	provider = newTestSQLiteProviderAtPath(t, path)

	require.NoError(t, provider.SchemaMigrate(ctx, false, schemaVersionEncryptionKeyDerivation))
	require.NoError(t, provider.Close())

	provider = newTestSQLiteProviderAtPath(t, path)

	require.NoError(t, provider.SchemaMigrate(ctx, false, schemaVersionEncryptionKeyDerivation-1))

	version, err := provider.SchemaVersion(ctx)

	require.NoError(t, err)
	assert.Equal(t, schemaVersionEncryptionKeyDerivation-1, version)

	result, err := provider.SchemaEncryptionCheckKey(ctx, true)

	require.NoError(t, err)
	assert.False(t, result.InvalidCheckValue)

	for table, tableResult := range result.Tables {
		assert.NoError(t, tableResult.Error, table)
		assert.Equal(t, 0, tableResult.Invalid, table)
	}

	require.NoError(t, provider.Close())

	provider = newTestSQLiteProviderAtPath(t, path)

	require.NoError(t, provider.SchemaMigrate(ctx, true, SchemaLatest))

	totp, err := provider.LoadTOTPConfiguration(ctx, "john")

	require.NoError(t, err)
	assert.Equal(t, totpSecret, totp.Secret)

	session, err := provider.LoadOAuth2Session(ctx, OAuth2SessionTypeAccessToken, "sig-123")

	require.NoError(t, err)
	assert.Equal(t, sessionData, session.Session)

	require.NoError(t, provider.Close())
}

func TestSchemaMigrateRollbackWithoutTx(t *testing.T) {
	totpSecret := []byte("JBSWY3DPEHPK3PXP")

	ctx := context.Background()

	provider := newTestSQLiteProvider(t)

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

	err := provider.schemaMigrateRollbackWithoutTx(ctx, schemaVersionEncryptionKeyDerivation-1, schemaVersionEncryptionAADRowScoped, errors.New("migration failed"))

	assert.EqualError(t, err, "migration rollback complete. rollback caused by: migration failed")

	version, err := provider.SchemaVersion(ctx)

	require.NoError(t, err)
	assert.Equal(t, schemaVersionEncryptionKeyDerivation-1, version)

	result, err := provider.SchemaEncryptionCheckKey(ctx, true)

	require.NoError(t, err)
	assert.False(t, result.InvalidCheckValue)

	for table, tableResult := range result.Tables {
		assert.NoError(t, tableResult.Error, table)
		assert.Equal(t, 0, tableResult.Invalid, table)
	}
}

func TestSchemaMigrateRollbackWithoutTxShouldErrOnUnknownMigrations(t *testing.T) {
	ctx := context.Background()

	provider := newTestSQLiteProvider(t)

	require.NoError(t, provider.SchemaMigrate(ctx, true, SchemaLatest))

	err := provider.schemaMigrateRollbackWithoutTx(ctx, SchemaLatest, SchemaLatest, errors.New("migration failed"))

	assert.EqualError(t, err, "error loading migrations from version 2147483647 to version 2147483647 for rollback: current version is same as migration target, no action being taken. rollback caused by: migration failed")
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

// saveLegacyOAuth2AccessTokenSession inserts an OAuth2.0 access token session using only the columns that predate
// the resource indicator columns added by a later migration. It exists so tests that pin the schema at an older
// version can still write a row without relying on provider.SaveOAuth2Session, which always targets the current
// (latest) column set.
func saveLegacyOAuth2AccessTokenSession(t *testing.T, ctx context.Context, provider *SQLiteProvider, session model.OAuth2Session) {
	t.Helper()

	var err error

	session.Session, err = utils.Encrypt(session.Session, provider.aad.Get(OAuth2SessionTypeAccessToken.AAD(), columnSessionData, session.Signature), provider.keys.encryption)
	require.NoError(t, err)

	query := fmt.Sprintf(`
		INSERT INTO %s (challenge_id, request_id, client_id, signature, subject, requested_at,
		requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, tableOAuth2AccessTokenSession)

	_, err = provider.db.ExecContext(ctx, query,
		session.ChallengeID, session.RequestID, session.ClientID, session.Signature,
		session.Subject, session.RequestedAt, session.RequestedScopes, session.GrantedScopes,
		session.RequestedAudience, session.GrantedAudience,
		session.Active, session.Revoked, session.Form, session.Session)
	require.NoError(t, err)
}

// loadLegacyOAuth2AccessTokenSessionData reads and decrypts the session_data column of an OAuth2.0 access token
// session directly, without going through provider.LoadOAuth2Session, whose SELECT always targets the current
// (latest) column set and so cannot be used against a schema pinned at an older version.
func loadLegacyOAuth2AccessTokenSessionData(t *testing.T, ctx context.Context, provider *SQLiteProvider, signature string) []byte {
	t.Helper()

	var encrypted []byte

	query := fmt.Sprintf(`SELECT session_data FROM %s WHERE signature = ?;`, tableOAuth2AccessTokenSession)

	require.NoError(t, provider.db.GetContext(ctx, &encrypted, query, signature))

	decrypted, err := utils.Decrypt(encrypted, provider.aad.Get(OAuth2SessionTypeAccessToken.AAD(), columnSessionData, signature), provider.keys.encryption)
	require.NoError(t, err)

	return decrypted
}
