package storage

import (
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
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
	dconfig := mysql.NewConfig()

	switch {
	case config.Port == 0:
		dconfig.Net = sqlNetworkTypeTCP
		dconfig.Addr = fmt.Sprintf("%s:%d", config.Host, 3306)
	default:
		dconfig.Net = sqlNetworkTypeTCP
		dconfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}

	switch config.Port {
	case 0:
		dconfig.Addr = config.Host
	default:
		dconfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}

	dconfig.DBName = config.Database
	dconfig.User = config.Username
	dconfig.Passwd = config.Password
	dconfig.Timeout = config.Timeout
	dconfig.MultiStatements = true
	dconfig.ParseTime = true
	dconfig.Loc = time.Local

	return dconfig.FormatDSN()
}
