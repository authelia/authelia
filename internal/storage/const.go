package storage

import (
	"regexp"
)

const (
	tableUserPreferences      = "user_preferences"
	tableIdentityVerification = "identity_verification"
	tableTOTPConfigurations   = "totp_configurations"
	tableWebauthnDevices      = "webauthn_devices"
	tableDuoDevices           = "duo_devices"
	tableAuthenticationLogs   = "authentication_logs"
	tableMigrations           = "migrations"
	tableEncryption           = "encryption"

	tablePrefixBackup = "_bkp_"
)

const (
	encryptionNameCheck = "check"
)

// WARNING: Do not change/remove these consts. They are used for Pre1 migrations.
const (
	tablePre1TOTPSecrets                = "totp_secrets"
	tablePre1IdentityVerificationTokens = "identity_verification_tokens"
	tablePre1U2FDevices                 = "u2f_devices"

	tablePre1Config = "config"

	tableAlphaAuthenticationLogs         = "AuthenticationLogs"
	tableAlphaIdentityVerificationTokens = "IdentityVerificationTokens"
	tableAlphaPreferences                = "Preferences"
	tableAlphaPreferencesTableName       = "PreferencesTableName"
	tableAlphaSecondFactorPreferences    = "SecondFactorPreferences"
	tableAlphaTOTPSecrets                = "TOTPSecrets"
	tableAlphaU2FDeviceHandles           = "U2FDeviceHandles"
)

var tablesPre1 = []string{
	tablePre1TOTPSecrets,
	tablePre1IdentityVerificationTokens,
	tablePre1U2FDevices,

	tableUserPreferences,
	tableAuthenticationLogs,
}

const (
	providerAll      = "all"
	providerMySQL    = "mysql"
	providerPostgres = "postgres"
	providerSQLite   = "sqlite"
)

const (
	// This is the latest schema version for the purpose of tests.
	testLatestVersion = 3
)

const (
	// SchemaLatest represents the value expected for a "migrate to latest" migration. It's the maximum 32bit signed integer.
	SchemaLatest = 2147483647
)

var (
	reMigration = regexp.MustCompile(`^V(\d{4})\.([^.]+)\.(all|sqlite|postgres|mysql)\.(up|down)\.sql$`)
)
