package storage

import (
	"crypto/x509"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// MSSQLProvider is a MySQL provider.
type MSSQLProvider struct {
	SQLProvider
}

// NewMSSQLProvider a MySQL provider.
func NewMSSQLProvider(config *schema.Configuration, caCertPool *x509.CertPool) (provider *MSSQLProvider) {
	provider = &MSSQLProvider{
		SQLProvider: NewSQLProvider(config, providerMSSQL, "sqlserver", config.Storage.MSSQL.Schema, dsnMSSQL(config.Storage.MSSQL, caCertPool)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMSSQLSelectExistingTables

	// Microsoft SQL requires rebinding of any query that contains a '?' placeholder.
	provider.rebind()

	return provider
}

func dsnMSSQL(config *schema.StorageMSSQL, caCertPool *x509.CertPool) (dataSourceName string) {
	fmt.Printf("creating dsn for %s and %d of db %s\n", config.Address.Hostname(), config.Address.Port(), config.Database)

	dsnConfig := msdsn.Config{
		Host:        config.Address.Hostname(),
		Port:        uint64(config.Address.Port()),
		Instance:    config.Instance,
		Database:    config.Database,
		User:        config.Username,
		Password:    config.Password,
		TLSConfig:   utils.NewTLSConfig(config.TLS, caCertPool),
		AppName:     driverParameterAppName,
		DialTimeout: config.Timeout,
		ConnTimeout: config.Timeout,
	}

	return dsnConfig.URL().String()
}
