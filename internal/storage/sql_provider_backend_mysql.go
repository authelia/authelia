package storage

import (
	"crypto/x509"
	"fmt"
	"strings"
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

	// MySQL uses ON DUPLICATE KEY UPDATE instead of ON CONFLICT.
	provider.sqlUpsertSession = fmt.Sprintf(queryFmtUpsertSessionMySQL, tableSessions)

	return provider
}

func dsnMySQL(config *schema.StorageMySQL, caCertPool *x509.CertPool) (dataSourceName string) {
	dsnConfig := mysql.NewConfig()

	dsnConfig.Net = config.Address.Network()
	dsnConfig.Addr = config.Address.NetworkAddress()

	if config.TLS != nil {
		dsnConfig.TLSConfig = fmt.Sprintf("authelia-%s-storage", utils.Version())

		_ = mysql.RegisterTLSConfig(dsnConfig.TLSConfig, utils.NewTLSConfig(config.TLS, caCertPool))
	}

	dsnConfig.DBName = config.Database
	dsnConfig.User = config.Username
	dsnConfig.Passwd = config.Password
	dsnConfig.Timeout = config.Timeout
	dsnConfig.MultiStatements = true
	dsnConfig.ParseTime = true
	dsnConfig.RejectReadOnly = true
	dsnConfig.Loc = time.Local
	dsnConfig.Collation = "utf8mb4_unicode_520_ci"
	dsnConfig.ConnectionAttributes = fmt.Sprintf("program_name:authelia,program_version:%s", strings.ReplaceAll(utils.Version(), ",", ""))

	return dsnConfig.FormatDSN()
}
