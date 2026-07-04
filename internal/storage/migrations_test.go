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

const (
	// This is the latest schema version for the purpose of tests.
	LatestVersion = 25
)

func TestShouldObtainCorrectMigrations(t *testing.T) {
	testCases := []struct {
		name     string
		provider string
	}{
		{
			"ShouldTestSQLite",
			providerSQLite,
		},
		{
			"ShouldTestPostgreSQL",
			providerPostgres,
		},
		{
			"ShouldTestMySQL",
			providerMySQL,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ver, err := latestMigrationVersion(tc.provider)
			require.NoError(t, err)

			assert.Equal(t, LatestVersion, ver)

			var (
				migrations []model.SchemaMigration
			)

			// UP.
			migrations, err = loadMigrations(tc.provider, 0, ver)
			require.NoError(t, err)

			assert.Len(t, migrations, ver)

			for i := 0; i < len(migrations); i++ {
				assert.Equal(t, i+1, migrations[i].Version)
			}

			migrations, err = loadMigrations(tc.provider, 1, ver)
			require.NoError(t, err)

			assert.Len(t, migrations, ver-1)

			// DOWN.
			migrations, err = loadMigrations(providerSQLite, ver, 0)
			require.NoError(t, err)

			assert.Len(t, migrations, ver)

			for i := 0; i < len(migrations); i++ {
				assert.Equal(t, ver-i, migrations[i].Version)
			}

			migrations, err = loadMigrations(tc.provider, ver, 1)
			require.NoError(t, err)

			assert.Len(t, migrations, ver-1)
		})
	}
}

func TestSchemaMigrateUpShouldMigrateWithoutStartupCheck(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	ctx := context.Background()

	require.NoError(t, provider.SchemaMigrate(ctx, true, SchemaLatest))

	version, err := provider.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, LatestVersion, version)

	assert.ErrorIs(t, provider.SchemaMigrate(ctx, true, SchemaLatest), ErrSchemaAlreadyUpToDate)

	history, err := provider.SchemaMigrationHistory(ctx)
	require.NoError(t, err)
	assert.Len(t, history, LatestVersion)

	up, err := provider.SchemaMigrationsUp(ctx, 0)
	require.NoError(t, err)
	assert.Empty(t, up)

	down, err := provider.SchemaMigrationsDown(ctx, 0)
	require.NoError(t, err)
	assert.Len(t, down, LatestVersion)

	result, err := provider.SchemaEncryptionCheckKey(ctx, false)
	require.NoError(t, err)
	assert.True(t, result.Success())
}

func TestSchemaMigrateDownShouldMigrateWithoutStartupCheck(t *testing.T) {
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
	require.NoError(t, migrator.StartupCheck())

	require.NoError(t, migrator.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
		CreatedAt: time.Now().Truncate(time.Second),
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    []byte("JBSWY3DPEHPK3PXP"),
	}))
	require.NoError(t, migrator.Close())

	provider, err := NewSQLiteProvider(config)
	require.NoError(t, err)

	require.NoError(t, provider.SchemaMigrate(ctx, false, 0))
}

func TestSchemaMigrateDownToZeroShouldSucceedWithStaleEncryptionKey(t *testing.T) {
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
	require.NoError(t, migrator.StartupCheck())

	require.NoError(t, migrator.SaveTOTPConfiguration(ctx, model.TOTPConfiguration{
		CreatedAt: time.Now().Truncate(time.Second),
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    []byte("JBSWY3DPEHPK3PXP"),
	}))

	require.NoError(t, migrator.SchemaEncryptionChangeKey(ctx, "authelia-new-test-key-not-a-secret-authelia-new-key"))
	require.NoError(t, migrator.Close())

	provider, err := NewSQLiteProvider(config)
	require.NoError(t, err)

	require.NoError(t, provider.SchemaMigrate(ctx, false, 0))
}

func TestMigrationShouldReturnErrorOnSame(t *testing.T) {
	migrations, err := loadMigrations(providerPostgres, 1, 1)

	assert.EqualError(t, err, "current version is same as migration target, no action being taken")
	assert.Nil(t, migrations)
}

func TestMigrationsShouldNotBeDuplicatedPostgres(t *testing.T) {
	migrations, err := loadMigrations(providerPostgres, 0, SchemaLatest)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(migrations))

	previousUp := make([]int, 0, len(migrations))

	for i, migration := range migrations {
		assert.True(t, migration.Up)

		if i != 0 {
			for _, v := range previousUp {
				assert.NotEqual(t, v, migration.Version)
			}
		}

		previousUp = append(previousUp, migration.Version)
	}

	migrations, err = loadMigrations(providerPostgres, SchemaLatest, 0)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(migrations))

	previousDown := make([]int, 0, len(migrations))

	for i, migration := range migrations {
		assert.False(t, migration.Up)

		if i != 0 {
			for _, v := range previousDown {
				assert.NotEqual(t, v, migration.Version)
			}
		}

		previousDown = append(previousDown, migration.Version)
	}
}

func TestMigrationsShouldNotBeDuplicatedMySQL(t *testing.T) {
	migrations, err := loadMigrations(providerMySQL, 0, SchemaLatest)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(migrations))

	previousUp := make([]int, 0, len(migrations))

	for i, migration := range migrations {
		assert.True(t, migration.Up)

		if i != 0 {
			for _, v := range previousUp {
				assert.NotEqual(t, v, migration.Version)
			}
		}

		previousUp = append(previousUp, migration.Version)
	}

	migrations, err = loadMigrations(providerMySQL, SchemaLatest, 0)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(migrations))

	previousDown := make([]int, 0, len(migrations))

	for i, migration := range migrations {
		assert.False(t, migration.Up)

		if i != 0 {
			for _, v := range previousDown {
				assert.NotEqual(t, v, migration.Version)
			}
		}

		previousDown = append(previousDown, migration.Version)
	}
}

func TestMigrationsShouldNotBeDuplicatedSQLite(t *testing.T) {
	migrations, err := loadMigrations(providerSQLite, 0, SchemaLatest)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(migrations))

	previousUp := make([]int, 0, len(migrations))

	for i, migration := range migrations {
		assert.True(t, migration.Up)

		if i != 0 {
			for _, v := range previousUp {
				assert.NotEqual(t, v, migration.Version)
			}
		}

		previousUp = append(previousUp, migration.Version)
	}

	migrations, err = loadMigrations(providerSQLite, SchemaLatest, 0)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(migrations))

	previousDown := make([]int, 0, len(migrations))

	for i, migration := range migrations {
		assert.False(t, migration.Up)

		if i != 0 {
			for _, v := range previousDown {
				assert.NotEqual(t, v, migration.Version)
			}
		}

		previousDown = append(previousDown, migration.Version)
	}
}
