package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnErrOnTargetSameAsCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks("sqlite", true, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks("sqlite", false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks("sqlite", false, 2, 2),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 2, 2))

	assert.EqualError(t,
		schemaMigrateChecks("mysql", false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks("postgres", false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))
}

func TestShouldReturnErrOnUpMigrationTargetVersionLessTHanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks("postgres", true, 1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 1, 2))

	assert.NoError(t,
		schemaMigrateChecks("postgres", true, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks("sqlite", true, 1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 1, 2))

	assert.NoError(t,
		schemaMigrateChecks("sqlite", true, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks("mysql", true, 1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 1, 2))

	assert.NoError(t,
		schemaMigrateChecks("mysql", true, 2, 1))
}

func TestMigrationUpShouldReturnErrOnAlreadyLatest(t *testing.T) {
	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks("postgres", true, SchemaLatest, 2))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks("mysql", true, SchemaLatest, 2))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks("sqlite", true, SchemaLatest, 2))
}

func TestShouldReturnErrOnVersionDoesntExits(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks("postgres", true, SchemaLatest-1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, 2))

	assert.EqualError(t,
		schemaMigrateChecks("mysql", true, SchemaLatest-1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, 2))

	assert.EqualError(t,
		schemaMigrateChecks("sqlite", true, SchemaLatest-1, 2),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, 2))
}

func TestMigrationDownShouldReturnErrOnTargetLessThanPre1(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks("sqlite", false, -4, 2),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -4))

	assert.EqualError(t,
		schemaMigrateChecks("mysql", false, -2, 2),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.EqualError(t,
		schemaMigrateChecks("postgres", false, -2, 2),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.NoError(t,
		schemaMigrateChecks("postgres", false, -1, 2))
}

func TestMigrationDownShouldReturnErrOnTargetVersionGreaterThanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks("sqlite", false, 2, 1),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks("mysql", false, 2, 1),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, 2, 1))

	assert.EqualError(t,
		schemaMigrateChecks("postgres", false, 2, 1),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, 2, 1))
}

func TestShouldReturnErrWhenCurrentIsGreaterThanLatest(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks("postgres", true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, 2))

	assert.EqualError(t,
		schemaMigrateChecks("mysql", true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, 2))

	assert.EqualError(t,
		schemaMigrateChecks("sqlite", true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, 2))
}

func TestSchemaVersionToString(t *testing.T) {
	assert.Equal(t, "unknown", SchemaVersionToString(-2))
	assert.Equal(t, "pre1", SchemaVersionToString(-1))
	assert.Equal(t, "N/A", SchemaVersionToString(0))
	assert.Equal(t, "1", SchemaVersionToString(1))
	assert.Equal(t, "2", SchemaVersionToString(2))
}
