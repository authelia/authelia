package storage

import (
	"errors"
)

var (
	// ErrNoAuthenticationLogs error thrown when no matching authentication logs have been found in DB.
	ErrNoAuthenticationLogs = errors.New("no matching authentication logs found")

	// ErrNoTOTPConfiguration error thrown when no TOTP configuration has been found in DB.
	ErrNoTOTPConfiguration = errors.New("no TOTP configuration for user")

	// ErrNoWebAuthnCredential error thrown when no WebAuthn credential handle has been found in DB.
	ErrNoWebAuthnCredential = errors.New("no WebAuthn credential found")

	// ErrNoDuoDevice error thrown when no Duo device and method has been found in DB.
	ErrNoDuoDevice = errors.New("no Duo device and method saved")

	// ErrNoAvailableMigrations is returned when no available migrations can be found.
	ErrNoAvailableMigrations = errors.New("no available migrations")

	// ErrMigrateCurrentVersionSameAsTarget is returned when the target version is the same as the current.
	ErrMigrateCurrentVersionSameAsTarget = errors.New("current version is same as migration target, no action being taken")

	// ErrSchemaAlreadyUpToDate is returned when the schema is already up to date.
	ErrSchemaAlreadyUpToDate = errors.New("schema already up to date")

	// ErrNoMigrationsFound is returned when no migrations were found.
	ErrNoMigrationsFound = errors.New("no schema migrations found")

	// ErrSchemaEncryptionVersionUnsupported is returned when the schema is checked if the encryption key is valid for
	// the database but the schema doesn't support encryption.
	ErrSchemaEncryptionVersionUnsupported = errors.New("schema version doesn't support encryption")

	// ErrSchemaEncryptionInvalidKey is returned when the schema is checked if the encryption key is valid for
	// the database but the key doesn't appear to be valid.
	ErrSchemaEncryptionInvalidKey = errors.New("the configured encryption key does not appear to be valid for this database which may occur if the encryption key was changed in the configuration without using the cli to change it in the database")
)

// Error formats for the storage provider.
const (
	ErrFmtMigrateUpTargetLessThanCurrent      = "schema up migration target version %d is less then the current version %d"
	ErrFmtMigrateUpTargetGreaterThanLatest    = "schema up migration target version %d is greater then the latest version %d which indicates it doesn't exist"
	ErrFmtMigrateDownTargetGreaterThanCurrent = "schema down migration target version %d is greater than the current version %d"
	ErrFmtMigrateDownTargetLessThanMinimum    = "schema down migration target version %d is less than the minimum version"
	ErrFmtMigrateAlreadyOnTargetVersion       = "schema migration target version %d is the same current version %d"
)

const (
	errFmtFailedMigration                     = "schema migration %d (%s) failed: %w"
	errFmtSchemaCurrentGreaterThanLatestKnown = "current schema version is greater than the latest known schema " +
		"version, you must downgrade to schema version %d before you can use this version of Authelia"
)

const (
	logFmtMigrationFromTo   = "Storage schema migration from %s to %s is being attempted"
	logFmtMigrationComplete = "Storage schema migration from %s to %s is complete"
	logFmtErrClosingConn    = "Error occurred closing SQL connection: %v"
)

const (
	errFmtMigrationPre1                 = "schema migration %s pre1 is no longer supported: you must use an older version of authelia to perform this migration: %s"
	errFmtMigrationPre1SuggestedVersion = "the suggested authelia version is 4.37.2"
)
