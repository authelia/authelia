package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// This is the latest schema version for the purpose of tests.
	LatestVersion = 13
)

func TestShouldObtainCorrectUpMigrations(t *testing.T) {
	ver, err := latestMigrationVersion(providerSQLite)
	require.NoError(t, err)

	assert.Equal(t, LatestVersion, ver)

	migrations, err := loadMigrations(providerSQLite, 0, ver)
	require.NoError(t, err)

	assert.Len(t, migrations, ver)

	for i := 0; i < len(migrations); i++ {
		assert.Equal(t, i+1, migrations[i].Version)
	}
}

func TestShouldObtainCorrectDownMigrations(t *testing.T) {
	ver, err := latestMigrationVersion(providerSQLite)
	require.NoError(t, err)

	assert.Equal(t, LatestVersion, ver)

	migrations, err := loadMigrations(providerSQLite, ver, 0)
	require.NoError(t, err)

	assert.Len(t, migrations, ver)

	for i := 0; i < len(migrations); i++ {
		assert.Equal(t, ver-i, migrations[i].Version)
	}
}

func TestMigrationShouldGetSpecificMigrationIfAvaliable(t *testing.T) {
	upMigrationsPostgreSQL, err := loadMigrations(providerPostgres, 8, 9)
	require.NoError(t, err)
	require.Len(t, upMigrationsPostgreSQL, 1)

	assert.True(t, upMigrationsPostgreSQL[0].Up)
	assert.Equal(t, 9, upMigrationsPostgreSQL[0].Version)
	assert.Equal(t, providerPostgres, upMigrationsPostgreSQL[0].Provider)

	upMigrationsSQLite, err := loadMigrations(providerSQLite, 8, 9)
	require.NoError(t, err)
	require.Len(t, upMigrationsSQLite, 1)

	assert.True(t, upMigrationsSQLite[0].Up)
	assert.Equal(t, 9, upMigrationsSQLite[0].Version)
	assert.Equal(t, providerAll, upMigrationsSQLite[0].Provider)

	downMigrationsPostgreSQL, err := loadMigrations(providerPostgres, 9, 8)
	require.NoError(t, err)
	require.Len(t, downMigrationsPostgreSQL, 1)

	assert.False(t, downMigrationsPostgreSQL[0].Up)
	assert.Equal(t, 9, downMigrationsPostgreSQL[0].Version)
	assert.Equal(t, providerAll, downMigrationsPostgreSQL[0].Provider)

	downMigrationsSQLite, err := loadMigrations(providerSQLite, 9, 8)
	require.NoError(t, err)
	require.Len(t, downMigrationsSQLite, 1)

	assert.False(t, downMigrationsSQLite[0].Up)
	assert.Equal(t, 9, downMigrationsSQLite[0].Version)
	assert.Equal(t, providerAll, downMigrationsSQLite[0].Provider)
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

	previousUp := make([]int, len(migrations))

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

	previousDown := make([]int, len(migrations))

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

	previousUp := make([]int, len(migrations))

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

	previousDown := make([]int, len(migrations))

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

	previousUp := make([]int, len(migrations))

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

	previousDown := make([]int, len(migrations))

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
