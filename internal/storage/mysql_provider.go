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
func NewMySQLProvider(config schema.MySQLStorageConfiguration, encryptionKey string) (provider *MySQLProvider) {
	provider = &MySQLProvider{
		SQLProvider: NewSQLProvider(providerMySQL, providerMySQL, dataSourceNameMySQL(config), encryptionKey),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMySQLSelectExistingTables

	// Specific alterations to this provider.
	provider.sqlFmtRenameTable = queryFmtMySQLRenameTable

	return provider
}

func dataSourceNameMySQL(config schema.MySQLStorageConfiguration) (dataSourceName string) {
	dataSourceName = config.Username

	if config.Password != "" {
		dataSourceName += fmt.Sprintf(":%s", config.Password)
	}

	if dataSourceName != "" {
		dataSourceName += "@"
	}

	address := config.Host
	if config.Port > 0 {
		address += fmt.Sprintf(":%d", config.Port)
	}

	dataSourceName += fmt.Sprintf("tcp(%s)", address)
	if config.Database != "" {
		dataSourceName += fmt.Sprintf("/%s", config.Database)
	}

	dataSourceName += "?"
	dataSourceName += fmt.Sprintf("timeout=%ds&multiStatements=true&parseTime=true", int32(config.Timeout/time.Second))

	return dataSourceName
}
