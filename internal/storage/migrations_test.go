package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldObtainCorrectUpMigrations(t *testing.T) {
	ver, err := latestMigrationVersion(provideerSQLite)
	require.NoError(t, err)

	assert.Equal(t, 1, ver)

	migrations, err := loadMigrations(provideerSQLite, 0, ver)
	require.NoError(t, err)

	assert.Len(t, migrations, ver)

	for i := 0; i < len(migrations); i++ {
		assert.Equal(t, i+1, migrations[i].Version)
	}
}

func TestShouldObtainCorrectDownMigrations(t *testing.T) {
	ver, err := latestMigrationVersion(provideerSQLite)
	require.NoError(t, err)

	assert.Equal(t, 1, ver)

	migrations, err := loadMigrations(provideerSQLite, ver, 0)
	require.NoError(t, err)

	assert.Len(t, migrations, ver)

	for i := 0; i < len(migrations); i++ {
		assert.Equal(t, ver-i, migrations[i].Version)
	}
}

func TestMigrationsShouldNotBeDuplicatedPostgres(t *testing.T) {
	migrations, err := loadMigrations("postgres", 0, SchemaLatest)
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

	migrations, err = loadMigrations("postgres", SchemaLatest, 0)
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
	migrations, err := loadMigrations("mysql", 0, SchemaLatest)
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

	migrations, err = loadMigrations("mysql", SchemaLatest, 0)
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
	migrations, err := loadMigrations("sqlite", 0, SchemaLatest)
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

	migrations, err = loadMigrations("sqlite", SchemaLatest, 0)
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
