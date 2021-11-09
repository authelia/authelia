package storage

import (
	"regexp"
)

const (
	tableUserPreferences      = "user_preferences"
	tableIdentityVerification = "identity_verification_tokens"
	tableTOTPConfigurations   = "totp_configurations"
	tableU2FDevices           = "u2f_devices"
	tableDUODevices           = "duo_devices"
	tableAuthenticationLogs   = "authentication_logs"
	tableMigrations           = "migrations"

	tablePrefixBackup = "_bkp_"
)

// WARNING: Do not change/remove these consts. They are used for Pre1 migrations.
const (
	tablePre1TOTPSecrets                 = "totp_secrets"
	tablePre1Config                      = "config"
	tablePre1IdentityVerificationTokens  = "identity_verification_tokens"
	tableAlphaAuthenticationLogs         = "AuthenticationLogs"
	tableAlphaIdentityVerificationTokens = "IdentityVerificationTokens"
	tableAlphaPreferences                = "Preferences"
	tableAlphaPreferencesTableName       = "PreferencesTableName"
	tableAlphaSecondFactorPreferences    = "SecondFactorPreferences"
	tableAlphaTOTPSecrets                = "TOTPSecrets"
	tableAlphaU2FDeviceHandles           = "U2FDeviceHandles"
)

var reMigration = regexp.MustCompile(`^V(\d{4})\.([^.]+)\.(all|sqlite|postgres|mysql)\.(up|down)\.sql$`)

const (
	providerAll      = "all"
	providerMySQL    = "mysql"
	providerPostgres = "postgres"
	provideerSQLite  = "sqlite"
)

const (
	SchemaLatest = 2147483647
)
