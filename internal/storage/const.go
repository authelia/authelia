package storage

import "fmt"

const storageSchemaCurrentVersion = 1
const storageSchemaUpgradeMessage = "Storage schema upgraded to version"
const storageSchemaUpgradeErrorText = "Storage schema upgrade failed at version"

// Keep table names in lower case because some DB does not support upper case.
const preferencesTableName = "user_preferences"
const identityVerificationTokensTableName = "identity_verification_tokens"
const totpSecretsTableName = "totp_secrets"
const u2fDeviceHandlesTableName = "u2f_devices"
const authenticationLogsTableName = "authentication_logs"
const configTableName = "config"

// SQLCreateUserPreferencesTable common SQL query to create user_preferences table.
var SQLCreateUserPreferencesTable = fmt.Sprintf(`
CREATE TABLE %s (
	username VARCHAR(100) PRIMARY KEY,
	second_factor_method VARCHAR(11)
)`, preferencesTableName)

// SQLCreateIdentityVerificationTokensTable common SQL query to create identity_verification_tokens table.
var SQLCreateIdentityVerificationTokensTable = fmt.Sprintf(`
CREATE TABLE %s (token VARCHAR(512))
`, identityVerificationTokensTableName)

// SQLCreateTOTPSecretsTable common SQL query to create totp_secrets table.
var SQLCreateTOTPSecretsTable = fmt.Sprintf(`
CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, secret VARCHAR(64))
`, totpSecretsTableName)

// SQLCreateU2FDeviceHandlesTable common SQL query to create u2f_device_handles table.
var SQLCreateU2FDeviceHandlesTable = fmt.Sprintf(`
CREATE TABLE %s (
	username VARCHAR(100) PRIMARY KEY,
	keyHandle TEXT,
	publicKey TEXT
)`, u2fDeviceHandlesTableName)

// SQLCreateAuthenticationLogsTable common SQL query to create authentication_logs table.
var SQLCreateAuthenticationLogsTable = fmt.Sprintf(`
CREATE TABLE %s (
	username VARCHAR(100),
	successful BOOL,
	time INTEGER
)`, authenticationLogsTableName)

// SQLCreateConfigTable common SQL query to create config table.
var SQLCreateConfigTable = fmt.Sprintf(`
CREATE TABLE %s (
	category VARCHAR(32) NOT NULL,
	key VARCHAR(32) NOT NULL,
    value TEXT,
	PRIMARY KEY (category, key)
)`, configTableName)
