// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"crypto/x509"
	"fmt"
	"regexp"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

var (
	reMSSQLBooleanTrue  = regexp.MustCompile(`\bTRUE\b`)
	reMSSQLBooleanFalse = regexp.MustCompile(`\bFALSE\b`)
	reMSSQLLength       = regexp.MustCompile(`\bLENGTH\(`)
)

func queryTransformMSSQLBitValues(query string) string {
	return reMSSQLBooleanFalse.ReplaceAllString(reMSSQLBooleanTrue.ReplaceAllString(query, "1"), "0")
}

func queryTransformMSSQLLength(query string) string {
	return reMSSQLLength.ReplaceAllString(query, "DATALENGTH(")
}

// MSSQLProvider is a Microsoft SQL provider.
type MSSQLProvider struct {
	SQLProvider
}

// NewMSSQLProvider a Microsoft SQL provider.
func NewMSSQLProvider(config *schema.Configuration, caCertPool *x509.CertPool) (provider *MSSQLProvider, err error) {
	var p SQLProvider

	if p, err = NewSQLProvider(config, providerMSSQL, "sqlserver", dsnMSSQL(config.Storage.MSSQL, caCertPool)); err != nil {
		return nil, err
	}

	provider = &MSSQLProvider{
		SQLProvider: p,
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryMSSQLSelectExistingTables

	// Microsoft SQL requires rebinding of any query that contains a '?' placeholder to use the '@p#' notation placeholders.
	provider.rebind(queryTransformMSSQLBitValues, queryTransformMSSQLLength)

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
	provider.sqlUpsertSession = fmt.Sprintf(queryFmtUpsertSessionMSSQL, tableSession)

	return provider, nil
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
