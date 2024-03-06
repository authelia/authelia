package storage

import (
	"crypto/x509"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// MySQLProvider is a MySQL provider.
type MySQLProvider struct {
	SQLProvider
}

// NewMySQLProvider a MySQL provider.
func NewMySQLProvider(config *schema.Configuration, caCertPool *x509.CertPool) (provider *MySQLProvider) {
	provider = &MySQLProvider{
		SQLProvider: NewSQLProvider(config, providerMySQL, providerMySQL, dsnMySQL(config.Storage.MySQL, caCertPool)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMySQLSelectExistingTables

	// Specific alterations to this provider.
	provider.sqlFmtRenameTable = queryFmtMySQLRenameTable

	return provider
}

func dsnMySQL(config *schema.StorageMySQL, caCertPool *x509.CertPool) (dataSourceName string) {
	dsnConfig := mysql.NewConfig()

	dsnConfig.Net = config.Address.Network()
	dsnConfig.Addr = config.Address.NetworkAddress()

	if config.TLS != nil {
		_ = mysql.RegisterTLSConfig("storage", utils.NewTLSConfig(config.TLS, caCertPool))

		dsnConfig.TLSConfig = "storage"
	}

	dsnConfig.DBName = config.Database
	dsnConfig.User = config.Username
	dsnConfig.Passwd = config.Password
	dsnConfig.Timeout = config.Timeout
	dsnConfig.MultiStatements = true
	dsnConfig.ParseTime = true
	dsnConfig.Loc = time.Local

	return dsnConfig.FormatDSN()
}
