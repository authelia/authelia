package storage

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/utils"
)

// schemaMigratePre1To1 takes the v1 migration and migrates to this version.
func (p *SQLProvider) schemaMigratePre1To1() (err error) {
	migration, err := loadMigration(p.name, 1, true)
	if err != nil {
		return err
	}

	// Get Tables list.
	tables, err := p.SchemaTables()
	if err != nil {
		return err
	}

	tablesRename := []string{
		tablePre1Config,
		tablePre1TOTPSecrets,
		tablePre1IdentityVerificationTokens,
		tableU2FDevices,
		tableUserPreferences,
		tableAuthenticationLogs,
		tableAlphaPreferences,
		tableAlphaIdentityVerificationTokens,
		tableAlphaAuthenticationLogs,
		tableAlphaPreferencesTableName,
		tableAlphaSecondFactorPreferences,
		tableAlphaTOTPSecrets,
		tableAlphaU2FDeviceHandles,
	}

	if err = p.schemaMigratePre1Rename(tables, tablesRename); err != nil {
		return err
	}

	if _, err = p.db.Exec(migration.Query); err != nil {
		return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
	}

	if _, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtPre1InsertUserPreferencesFromSelect),
		tableUserPreferences, tablePrefixBackup+tableUserPreferences)); err != nil {
		return err
	}

	if err = p.schemaMigratePre1To1AuthenticationLogs(); err != nil {
		return err
	}

	if err = p.schemaMigratePre1To1U2F(); err != nil {
		return err
	}

	if err = p.schemaMigratePre1To1TOTP(); err != nil {
		return err
	}

	for _, table := range tablesRename {
		if _, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtDropTableIfExists), tablePrefixBackup+table)); err != nil {
			return err
		}
	}

	return p.schemaMigrateFinalizeAdvanced(-1, 1)
}

func (p *SQLProvider) schemaMigratePre1Rename(tables, tablesRename []string) (err error) {
	// Rename Tables and Indexes.
	for _, table := range tables {
		if !utils.IsStringInSlice(table, tablesRename) {
			continue
		}

		tableNew := tablePrefixBackup + table

		if _, err = p.db.Exec(fmt.Sprintf(p.sqlFmtRenameTable, table, tableNew)); err != nil {
			return err
		}

		if p.name == providerPostgres {
			if table == tableU2FDevices || table == tableUserPreferences {
				if _, err = p.db.Exec(fmt.Sprintf(`ALTER TABLE %s RENAME CONSTRAINT %s_pkey TO %s_pkey;`,
					tableNew, table, tableNew)); err != nil {
					continue
				}
			}

			if _, err = p.db.Exec(`ALTER INDEX IF EXISTS authentication_logs_username_idx RENAME TO _bkp_authentication_logs_username_idx;`); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigratePre1To1Rollback(up bool) (err error) {
	if up {
		migration, err := loadMigration(p.name, 1, false)
		if err != nil {
			return err
		}

		if _, err = p.db.Exec(migration.Query); err != nil {
			return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
		}
	}

	tables, err := p.SchemaTables()
	if err != nil {
		return err
	}

	for _, table := range tables {
		if !strings.HasPrefix(table, tablePrefixBackup) {
			continue
		}

		tableNew := strings.Replace(table, tablePrefixBackup, "", 1)
		if _, err = p.db.Exec(fmt.Sprintf(p.sqlFmtRenameTable, table, tableNew)); err != nil {
			return err
		}

		if p.name == providerPostgres && (tableNew == tableU2FDevices || tableNew == tableUserPreferences) {
			if _, err = p.db.Exec(fmt.Sprintf(`ALTER TABLE %s RENAME CONSTRAINT %s_pkey TO %s_pkey;`,
				tableNew, table, tableNew)); err != nil {
				continue
			}
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigratePre1To1AuthenticationLogs() (err error) {
	for page := 0; true; page++ {
		attempts, err := p.schemaMigratePre1To1AuthenticationLogsGetRows(page)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}

			return err
		}

		for _, attempt := range attempts {
			_, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtPre1To1InsertAuthenticationLogs), tableAuthenticationLogs), attempt.Username, attempt.Successful, attempt.Time)
			if err != nil {
				return err
			}
		}

		if len(attempts) != 100 {
			break
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigratePre1To1AuthenticationLogsGetRows(page int) (attempts []models.AuthenticationAttempt, err error) {
	rows, err := p.db.Queryx(fmt.Sprintf(p.db.Rebind(queryFmtPre1To1SelectAuthenticationLogs), tablePrefixBackup+tableAuthenticationLogs), page*100)
	if err != nil {
		return nil, err
	}

	attempts = make([]models.AuthenticationAttempt, 0, 100)

	for rows.Next() {
		var (
			username   string
			successful bool
			timestamp  int64
		)

		err = rows.Scan(&username, &successful, &timestamp)
		if err != nil {
			return nil, err
		}

		attempts = append(attempts, models.AuthenticationAttempt{Username: username, Successful: successful, Time: time.Unix(timestamp, 0)})
	}

	return attempts, nil
}

func (p *SQLProvider) schemaMigratePre1To1TOTP() (err error) {
	rows, err := p.db.Queryx(fmt.Sprintf(p.db.Rebind(queryFmtPre1SelectTOTPConfigurations), tablePrefixBackup+tablePre1TOTPSecrets))
	if err != nil {
		return err
	}

	var totpConfigs []models.TOTPConfiguration

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf(logFmtErrClosingConn, err)
		}
	}()

	for rows.Next() {
		var username, secret string

		err = rows.Scan(&username, &secret)
		if err != nil {
			return err
		}

		// TODO: Add encryption migration here.
		encryptedSecret := "encrypted:" + secret

		totpConfigs = append(totpConfigs, models.TOTPConfiguration{Username: username, Secret: encryptedSecret})
	}

	for _, config := range totpConfigs {
		_, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtPre1InsertTOTPConfiguration), tableTOTPConfigurations), config.Username, config.Secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigratePre1To1U2F() (err error) {
	rows, err := p.db.Queryx(fmt.Sprintf(p.db.Rebind(queryFmtPre1To1SelectU2FDevices), tablePrefixBackup+tableU2FDevices))
	if err != nil {
		return err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf(logFmtErrClosingConn, err)
		}
	}()

	var devices []models.U2FDevice

	for rows.Next() {
		var username, keyHandleBase64, publicKeyBase64 string

		err = rows.Scan(&username, &keyHandleBase64, &publicKeyBase64)
		if err != nil {
			return err
		}

		keyHandle, err := base64.StdEncoding.DecodeString(keyHandleBase64)
		if err != nil {
			return err
		}

		publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
		if err != nil {
			return err
		}

		devices = append(devices, models.U2FDevice{Username: username, KeyHandle: keyHandle, PublicKey: publicKey})
	}

	for _, device := range devices {
		_, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtPre1To1InsertU2FDevice), tableU2FDevices), device.Username, device.KeyHandle, device.PublicKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate1ToPre1() (err error) {
	tables, err := p.SchemaTables()
	if err != nil {
		return err
	}

	tablesRename := []string{
		tableMigrations,
		tableTOTPConfigurations,
		tableIdentityVerification,
		tableU2FDevices,
		tableDUODevices,
		tableUserPreferences,
		tableAuthenticationLogs,
	}

	if err = p.schemaMigratePre1Rename(tables, tablesRename); err != nil {
		return err
	}

	if _, err := p.db.Exec(queryCreatePre1); err != nil {
		return err
	}

	if _, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtPre1InsertUserPreferencesFromSelect),
		tableUserPreferences, tablePrefixBackup+tableUserPreferences)); err != nil {
		return err
	}

	if err = p.schemaMigrate1ToPre1AuthenticationLogs(); err != nil {
		return err
	}

	if err = p.schemaMigrate1ToPre1U2F(); err != nil {
		return err
	}

	if err = p.schemaMigrate1ToPre1TOTP(); err != nil {
		return err
	}

	queryFmtDropTableRebound := p.db.Rebind(queryFmtDropTableIfExists)

	for _, table := range tablesRename {
		if _, err = p.db.Exec(fmt.Sprintf(queryFmtDropTableRebound, tablePrefixBackup+table)); err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate1ToPre1AuthenticationLogs() (err error) {
	for page := 0; true; page++ {
		attempts, err := p.schemaMigrate1ToPre1AuthenticationLogsGetRows(page)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}

			return err
		}

		for _, attempt := range attempts {
			_, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1InsertAuthenticationLogs), tableAuthenticationLogs), attempt.Username, attempt.Successful, attempt.Time.Unix())
			if err != nil {
				return err
			}
		}

		if len(attempts) != 100 {
			break
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate1ToPre1AuthenticationLogsGetRows(page int) (attempts []models.AuthenticationAttempt, err error) {
	rows, err := p.db.Queryx(fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1SelectAuthenticationLogs), tablePrefixBackup+tableAuthenticationLogs), page*100)
	if err != nil {
		return nil, err
	}

	attempts = make([]models.AuthenticationAttempt, 0, 100)

	var attempt models.AuthenticationAttempt
	for rows.Next() {
		err = rows.StructScan(&attempt)
		if err != nil {
			return nil, err
		}

		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

func (p *SQLProvider) schemaMigrate1ToPre1TOTP() (err error) {
	rows, err := p.db.Queryx(fmt.Sprintf(p.db.Rebind(queryFmtPre1SelectTOTPConfigurations), tablePrefixBackup+tableTOTPConfigurations))
	if err != nil {
		return err
	}

	var totpConfigs []models.TOTPConfiguration

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf(logFmtErrClosingConn, err)
		}
	}()

	for rows.Next() {
		var username, encryptedSecret string

		err = rows.Scan(&username, &encryptedSecret)
		if err != nil {
			return err
		}

		// TODO: Fix.
		// TODO: Add DECRYPTION migration here.
		decryptedSecret := strings.Replace(encryptedSecret, "encrypted:", "", 1)

		totpConfigs = append(totpConfigs, models.TOTPConfiguration{Username: username, Secret: decryptedSecret})
	}

	for _, config := range totpConfigs {
		_, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtPre1InsertTOTPConfiguration), tablePre1TOTPSecrets), config.Username, config.Secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate1ToPre1U2F() (err error) {
	rows, err := p.db.Queryx(fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1SelectU2FDevices), tablePrefixBackup+tableU2FDevices))
	if err != nil {
		return err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf(logFmtErrClosingConn, err)
		}
	}()

	var (
		devices []models.U2FDevice
		device  models.U2FDevice
	)

	for rows.Next() {
		err = rows.StructScan(&device)
		if err != nil {
			return err
		}

		devices = append(devices, device)
	}

	for _, device := range devices {
		_, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1InsertU2FDevice), tableU2FDevices), device.Username, base64.StdEncoding.EncodeToString(device.KeyHandle), base64.StdEncoding.EncodeToString(device.PublicKey))
		if err != nil {
			return err
		}
	}

	return nil
}
