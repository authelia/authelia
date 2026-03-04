package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

const (
	// This is the latest schema version for the purpose of tests.
	LatestVersion = 24
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
