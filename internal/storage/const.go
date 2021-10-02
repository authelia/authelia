package storage

const storageSchemaCurrentVersion = SchemaVersion(1)
const storageSchemaUpgradeMessage = "Storage schema upgraded to v"
const storageSchemaUpgradeErrorText = "storage schema upgrade failed at v"

// Keep table names in lower case because some DB does not support upper case.
const (
	tableUserPreferences            = "user_preferences"
	tableIdentityVerificationTokens = "identity_verification_tokens"
	tableTOTPSecrets                = "totp_secrets"
	tableU2FDevices                 = "u2f_devices"
	tableAuthenticationLogs         = "authentication_logs"
	tableMigrations                 = "migrations"
	tableConfig                     = "config"
)
