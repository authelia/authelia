// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"crypto/x509"
	"fmt"
	"path"
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

func dsnMySQL(config *schema.MySQLStorageConfiguration, caCertPool *x509.CertPool) (dataSourceName string) {
	dsnConfig := mysql.NewConfig()

	switch {
	case path.IsAbs(config.Host):
		dsnConfig.Net = sqlNetworkTypeUnixSocket
		dsnConfig.Addr = config.Host
	case config.Port == 0:
		dsnConfig.Net = sqlNetworkTypeTCP
		dsnConfig.Addr = fmt.Sprintf("%s:%d", config.Host, 3306)
	default:
		dsnConfig.Net = sqlNetworkTypeTCP
		dsnConfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}

	if config.TLS != nil {
		_ = mysql.RegisterTLSConfig("storage", utils.NewTLSConfig(config.TLS, caCertPool))

		dsnConfig.TLSConfig = "storage"
	}

	switch config.Port {
	case 0:
		dsnConfig.Addr = config.Host
	default:
		dsnConfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
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
