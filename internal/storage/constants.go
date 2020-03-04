package storage

import "fmt"

// Keep table names in lower case because some DB does not support upper case.
var preferencesTableName = "user_preferences"
var identityVerificationTokensTableName = "identity_verification_tokens"
var totpSecretsTableName = "totp_secrets"
var u2fDeviceHandlesTableName = "u2f_devices"
var authenticationLogsTableName = "authentication_logs"

// SQLCreateUserPreferencesTable common SQL query to create user_preferences table
var SQLCreateUserPreferencesTable = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	username VARCHAR(100) PRIMARY KEY,
	second_factor_method VARCHAR(11)
)`, preferencesTableName)

// SQLCreateIdentityVerificationTokensTable common SQL query to create identity_verification_tokens table
var SQLCreateIdentityVerificationTokensTable = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (token VARCHAR(512))
`, identityVerificationTokensTableName)

// SQLCreateTOTPSecretsTable common SQL query to create totp_secrets table
var SQLCreateTOTPSecretsTable = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (username VARCHAR(100) PRIMARY KEY, secret VARCHAR(64))
`, totpSecretsTableName)

// SQLCreateU2FDeviceHandlesTable common SQL query to create u2f_device_handles table
var SQLCreateU2FDeviceHandlesTable = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	username VARCHAR(100) PRIMARY KEY,
	keyHandle TEXT,
	publicKey TEXT
)`, u2fDeviceHandlesTableName)

// SQLCreateAuthenticationLogsTable common SQL query to create authentication_logs table
var SQLCreateAuthenticationLogsTable = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	username VARCHAR(100),
	successful BOOL,
	time INTEGER,
	INDEX usr_time_idx (username, time)
)`, authenticationLogsTableName)
