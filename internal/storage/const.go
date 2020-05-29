package storage

import (
	"fmt"
)

const storageSchemaCurrentVersion = 2
const storageSchemaUpgradeMessage = "Storage schema upgraded to v"
const storageSchemaUpgradeErrorText = "storage schema upgrade failed at v"

// Keep table names in lower case because some DB does not support upper case.
const userPreferencesTableName = "user_preferences"
const identityVerificationTokensTableName = "identity_verification_tokens"
const totpSecretsTableName = "totp_secrets"
const u2fDeviceHandlesTableName = "u2f_devices"
const authenticationLogsTableName = "authentication_logs"
const configTableName = "config"

/*
// SQLCreateUserPreferencesTable common SQL query to create user_preferences table.
var SQLCreateUserPreferencesTable = fmt.Sprintf(`
CREATE TABLE %s (
	username VARCHAR(100) PRIMARY KEY,
	second_factor_method VARCHAR(11)
)`, userPreferencesTableName)

// SQLCreateIdentityVerificationTokensTable common SQL query to create identity_verification_tokens table.
var SQLCreateIdentityVerificationTokensTable = fmt.Sprintf(`
CREATE TABLE %s (
	token VARCHAR(512)
)`, identityVerificationTokensTableName)

// SQLCreateTOTPSecretsTable common SQL query to create totp_secrets table.
var SQLCreateTOTPSecretsTable = fmt.Sprintf(`
CREATE TABLE %s (
	username VARCHAR(100) PRIMARY KEY,
	secret VARCHAR(64)
)`, totpSecretsTableName)

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
	key_name VARCHAR(32) NOT NULL,
    value TEXT,
	PRIMARY KEY (category, key_name)
)`, configTableName)
.
*/

// sqlUpgradeCreateTableStatements is a map of the schema version number, plus a map of the table name and the statement used to create it.
// The statement is fmt.Sprintf'd with the table name as the first argument.
var sqlUpgradeCreateTableStatements = map[int]map[string]string{
	1: {
		userPreferencesTableName:            "CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, second_factor_method VARCHAR(11))",
		identityVerificationTokensTableName: "CREATE TABLE %s (token VARCHAR(512))",
		totpSecretsTableName:                "CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, secret VARCHAR(64))",
		u2fDeviceHandlesTableName:           "CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, keyHandle TEXT, publicKey TEXT)",
		authenticationLogsTableName:         "CREATE TABLE %s (username VARCHAR(100), successful BOOL, time INTEGER)",
		configTableName:                     "CREATE TABLE %s (category VARCHAR(32) NOT NULL, key_name VARCHAR(32) NOT NULL, value TEXT, PRIMARY KEY (category, key_name))",
	},
}

// sqlUpgradesCreateTableIndexesStatements is a map of t he schema version number, plus a slice of statements to create all of the indexes.
var sqlUpgradesCreateTableIndexesStatements = map[int][]string{
	1: {
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s (username, time)", authenticationLogsTableName),
	},
}
