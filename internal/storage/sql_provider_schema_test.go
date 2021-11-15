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

func TestShouldReturnErrOnUpMigrationTargetVersionLessTHanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, 1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 1, 2))

	assert.NoError(t,
		schemaMigrateChecks(providerPostgres, true, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, 1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 1, 2))

	assert.NoError(t,
		schemaMigrateChecks(providerSQLite, true, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, 1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 1, 2))

	assert.NoError(t,
		schemaMigrateChecks(providerMySQL, true, 2, 1))
}

func TestMigrationUpShouldReturnErrOnAlreadyLatest(t *testing.T) {
	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest, testLatestVersion))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest, testLatestVersion))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest, testLatestVersion))
}

func TestShouldReturnErrOnVersionDoesntExits(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest-1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest-1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest-1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, testLatestVersion))
}

func TestMigrationDownShouldReturnErrOnTargetLessThanPre1(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, -4, 2),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -4))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, -2, 2),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, -2, 2),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.NoError(t,
		schemaMigrateChecks(providerPostgres, false, -1, 2))
}

func TestMigrationDownShouldReturnErrOnTargetVersionGreaterThanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, 2, 1),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, 2, 1),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, 2, 1),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, 2, 1))
}

func TestShouldReturnErrWhenCurrentIsGreaterThanLatest(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, testLatestVersion))
}

func TestSchemaVersionToString(t *testing.T) {
	assert.Equal(t, "unknown", SchemaVersionToString(-2))
	assert.Equal(t, "pre1", SchemaVersionToString(-1))
	assert.Equal(t, "N/A", SchemaVersionToString(0))
	assert.Equal(t, "1", SchemaVersionToString(1))
	assert.Equal(t, "2", SchemaVersionToString(2))
}
