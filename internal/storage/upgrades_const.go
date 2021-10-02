package storage

import (
	"fmt"
)

// sqlUpgradeCreateTableStatements is a map of the schema version number, plus a map of the table name and the statement used to create it.
// The statement is fmt.Sprintf'd with the table name as the first argument.
var sqlUpgradeCreateTableStatements = map[SchemaVersion]map[string]string{
	SchemaVersion(1): {
		tableUserPreferences: `
CREATE TABLE %s (
	username VARCHAR(100) PRIMARY KEY, 
	second_factor_method VARCHAR(11)
);`,

		tableIdentityVerificationTokens: `
CREATE TABLE %s (
	token VARCHAR(512)
);`,

		tableTOTPSecrets: `
CREATE TABLE %s (
	username VARCHAR(100) PRIMARY KEY,
	secret VARCHAR(64)
);`,

		tableU2FDevices: `
CREATE TABLE %s (
	username VARCHAR(100) PRIMARY KEY,
	key_handle BLOB,
	public_key BLOB
);`,

		tableAuthenticationLogs: `
CREATE TABLE %s (
	username VARCHAR(100),
	successful BOOL,
	time INTEGER
);`,
	},
}

// sqlUpgradesCreateTableIndexesStatements is a map of t he schema version number, plus a slice of statements to create all of the indexes.
var sqlUpgradesCreateTableIndexesStatements = map[SchemaVersion][]string{
	SchemaVersion(1): {
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s (username, time)", tableAuthenticationLogs),
	},
}
