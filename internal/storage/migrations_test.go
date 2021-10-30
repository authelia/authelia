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
