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
		SQLProvider: NewSQLProvider(config, providerMSSQL, "sqlserver", "", dsnMSSQL(config.Storage.MSSQL, caCertPool)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMSSQLSelectExistingTables

	// Specific alterations to this provider.
	// MSSQL doesn't have a UPSERT statement but has an ON DUPLICATE KEY operation instead.
	provider.sqlUpsertDuoDevice = fmt.Sprintf(queryFmtUpsertDuoDevicePostgreSQL, tableDuoDevices)
	provider.sqlUpsertTOTPConfig = fmt.Sprintf(queryFmtUpsertTOTPConfigurationPostgreSQL, tableTOTPConfigurations)
	provider.sqlUpsertPreferred2FAMethod = fmt.Sprintf(queryFmtUpsertPreferred2FAMethodPostgreSQL, tableUserPreferences)
	provider.sqlUpsertEncryptionValue = fmt.Sprintf(queryFmtUpsertEncryptionValueMSSQL, tableEncryption, tableEncryption)
	provider.sqlUpsertOAuth2BlacklistedJTI = fmt.Sprintf(queryFmtUpsertOAuth2BlacklistedJTIPostgreSQL, tableOAuth2BlacklistedJTI)
	provider.sqlInsertOAuth2ConsentPreConfiguration = fmt.Sprintf(queryFmtInsertOAuth2ConsentPreConfigurationPostgreSQL, tableOAuth2ConsentPreConfiguration)
	provider.sqlUpsertCachedData = fmt.Sprintf(queryFmtUpsertCachedDataPostgreSQL, tableCachedData)

	// Microsoft SQL requires rebinding of any query that contains a '?' placeholder.
	provider.rebind()

	return provider
}

func dsnMSSQL(config *schema.StorageMSSQL, caCertPool *x509.CertPool) (dataSourceName string) {
	var encryption msdsn.Encryption

	if config.TLS != nil {
		encryption = msdsn.EncryptionStrict
	}

	dsnConfig := msdsn.Config{
		Port:        uint64(config.Address.Port()),
		Host:        config.Address.Hostname(),
		Instance:    config.Instance,
		Database:    config.Database,
		User:        config.Username,
		Password:    config.Password,
		Encryption:  encryption,
		TLSConfig:   utils.NewTLSConfig(config.TLS, caCertPool),
		AppName:     fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		DialTimeout: config.Timeout,
		ConnTimeout: config.Timeout,
	}

	return dsnConfig.URL().String()
}
