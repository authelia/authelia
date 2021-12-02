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
func NewMySQLProvider(config *schema.Configuration) (provider *MySQLProvider) {
	provider = &MySQLProvider{
		SQLProvider: NewSQLProvider(config, providerMySQL, providerMySQL, dataSourceNameMySQL(*config.Storage.MySQL)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMySQLSelectExistingTables

	// Specific alterations to this provider.
	provider.sqlFmtRenameTable = queryFmtMySQLRenameTable

	return provider
}

func dataSourceNameMySQL(config schema.MySQLStorageConfiguration) (dataSourceName string) {
	dataSourceName = fmt.Sprintf("%s:%s", config.Username, config.Password)

	if dataSourceName != "" {
		dataSourceName += "@"
	}

	address := config.Host
	if config.Port > 0 {
		address += fmt.Sprintf(":%d", config.Port)
	}

	dataSourceName += fmt.Sprintf("tcp(%s)/%s", address, config.Database)

	dataSourceName += "?"
	dataSourceName += fmt.Sprintf("timeout=%ds&multiStatements=true&parseTime=true", int32(config.Timeout/time.Second))

	return dataSourceName
}
