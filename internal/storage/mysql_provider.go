package storage

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // Load the MySQL Driver used in the connection string.

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// MySQLProvider is a MySQL provider.
type MySQLProvider struct {
	SQLProvider
}

// NewMySQLProvider a MySQL provider.
func NewMySQLProvider(config schema.MySQLStorageConfiguration) (provider *MySQLProvider) {
	provider = &MySQLProvider{
		SQLProvider: NewSQLProvider("mysql", "mysql", buildMySQLConnectionString(config)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMySQLSelectExistingTables

	// Specific alterations to this provider.
	provider.sqlRenameTable = queryMySQLRenameTable

	// TODO: Remove this as part of the migrations change.
	provider.sqlUpgradesCreateTableStatements[SchemaVersion(1)][tableAuthenticationLogs] = "CREATE TABLE %s (username VARCHAR(100), successful BOOL, time INTEGER, INDEX usr_time_idx (username, time))"
	provider.sqlConfigSetValue = fmt.Sprintf("REPLACE INTO %s (category, key_name, value) VALUES (?, ?, ?)", tableConfig)
	provider.sqlConfigGetValue = fmt.Sprintf("SELECT value FROM %s WHERE category=? AND key_name=?", tableConfig)

	return provider
}

func buildMySQLConnectionString(config schema.MySQLStorageConfiguration) (connectionString string) {
	connectionString = config.Username

	if config.Password != "" {
		connectionString += fmt.Sprintf(":%s", config.Password)
	}

	if connectionString != "" {
		connectionString += "@"
	}

	address := config.Host
	if config.Port > 0 {
		address += fmt.Sprintf(":%d", config.Port)
	}

	connectionString += fmt.Sprintf("tcp(%s)", address)
	if config.Database != "" {
		connectionString += fmt.Sprintf("/%s", config.Database)
	}

	connectionString += "?"
	connectionString += fmt.Sprintf("timeout=%ds", int32(config.Timeout/time.Second))

	return connectionString
}
