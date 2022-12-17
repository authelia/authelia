package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
