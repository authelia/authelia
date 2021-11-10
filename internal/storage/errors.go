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

	// ErrUnknownSchemaState is returned when the schema state is unknown.
	ErrUnknownSchemaState = errors.New("unknown schema state")
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
