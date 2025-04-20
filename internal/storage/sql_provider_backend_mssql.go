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
		SQLProvider: NewSQLProvider(config, providerMSSQL, "sqlserver", dsnMSSQL(config.Storage.MSSQL, caCertPool)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMSSQLSelectExistingTables

	// Microsoft SQL requires rebinding of any query that contains a '?' placeholder to use the '@p#' notation placeholders.
	provider.rebind()

	/*
		Specific query adjustments for this provider.
	*/

	// Microsoft SQL doesn't have a UPSERT statement but has TRY/CATCH logic instead.
	provider.sqlUpsertDuoDevice = fmt.Sprintf(queryFmtUpsertDuoDeviceMSSQL, tableDuoDevices, tableDuoDevices)
	provider.sqlUpsertTOTPConfig = fmt.Sprintf(queryFmtUpsertTOTPConfigurationMSSQL, tableTOTPConfigurations, tableTOTPConfigurations)
	provider.sqlUpsertPreferred2FAMethod = fmt.Sprintf(queryFmtUpsertPreferred2FAMethodMSSQL, tableUserPreferences, tableUserPreferences)
	provider.sqlUpsertEncryptionValue = fmt.Sprintf(queryFmtUpsertEncryptionValueMSSQL, tableEncryption, tableEncryption)
	provider.sqlUpsertOAuth2BlacklistedJTI = fmt.Sprintf(queryFmtUpsertOAuth2BlacklistedJTIMSSQL, tableOAuth2BlacklistedJTI, tableOAuth2BlacklistedJTI)
	provider.sqlUpsertCachedData = fmt.Sprintf(queryFmtUpsertCachedDataMSSQL, tableCachedData, tableCachedData)

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
