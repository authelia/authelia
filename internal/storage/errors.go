package storage

import (
	"errors"
)

var (
	// ErrNoU2FDeviceHandle error thrown when no U2F device handle has been found in DB.
	ErrNoU2FDeviceHandle = errors.New("no U2F device handle found")

	// ErrNoTOTPSecret error thrown when no TOTP secret has been found in DB.
	ErrNoTOTPSecret = errors.New("no TOTP secret registered")

	// ErrNoAvailableMigrations is returned when no available migrations can be found.
	ErrNoAvailableMigrations = errors.New("no available migrations")

	// ErrSchemaAlreadyUpToDate is returned when the schema is already up to date.
	ErrSchemaAlreadyUpToDate = errors.New("schema already up to date")
)

// Error formats for the storage provider.
const (
	ErrFmtMigrateUpTargetLessThanCurrent      = "schema up migration target version %d is less then the current version %d"
	ErrFmtMigrateUpTargetGreaterThanLatest    = "schema up migration target version %d is less then the latest version %d"
	ErrFmtMigrateDownTargetGreaterThanCurrent = "schema down migration target version %d is greater than the current version %d"
	ErrFmtMigrateDownTargetLessThanMinimum    = "schema down migration target version %d is less than the minimum version"
	ErrFmtMigrateAlreadyOnTargetVersion       = "schema migration target version %d is the same current version %d"
)

const (
	errFmtFailedMigration                     = "schema migration %d (%s) failed: %w"
	errFmtFailedMigrationPre1                 = "schema migration pre1 failed: %w"
	errFmtSchemaCurrentGreaterThanLatestKnown = "current schema version is greater than the latest known schema " +
		"version, you must downgrade to schema version %d before you can use this version of Authelia"
)

const (
	logFmtMigrationFromTo   = "Storage schema migration from %s to %s is being attempted"
	logFmtMigrationComplete = "Storage schema migration from %s to %s is complete"
	logFmtErrClosingConn    = "Error occurred closing SQL connection: %v"
)
