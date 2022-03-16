package storage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

// schemaMigratePre1To1 takes the v1 migration and migrates to this version.
func (p *SQLProvider) schemaMigratePre1To1(ctx context.Context) (err error) {
	migration, err := loadMigration(p.name, 1, true)
	if err != nil {
		return err
	}

	// Get Tables list.
	tables, err := p.SchemaTables(ctx)
	if err != nil {
		return err
	}

	tablesRename := []string{
		tablePre1Config,
		tablePre1TOTPSecrets,
		tablePre1IdentityVerificationTokens,
		tablePre1U2FDevices,
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

	if err = p.schemaMigratePre1Rename(ctx, tables, tablesRename); err != nil {
		return err
	}

	if _, err = p.db.ExecContext(ctx, migration.Query); err != nil {
		return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
	}

	if err = p.setNewEncryptionCheckValue(ctx, &p.key, nil); err != nil {
		return err
	}

	if _, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1InsertUserPreferencesFromSelect),
		tableUserPreferences, tablePrefixBackup+tableUserPreferences)); err != nil {
		return err
	}

	if err = p.schemaMigratePre1To1AuthenticationLogs(ctx); err != nil {
		return err
	}

	if err = p.schemaMigratePre1To1U2F(ctx); err != nil {
		return err
	}

	if err = p.schemaMigratePre1To1TOTP(ctx); err != nil {
		return err
	}

	for _, table := range tablesRename {
		if _, err = p.db.Exec(fmt.Sprintf(p.db.Rebind(queryFmtDropTableIfExists), tablePrefixBackup+table)); err != nil {
			return err
		}
	}

	return p.schemaMigrateFinalizeAdvanced(ctx, -1, 1)
}

func (p *SQLProvider) schemaMigratePre1Rename(ctx context.Context, tables, tablesRename []string) (err error) {
	// Rename Tables and Indexes.
	for _, table := range tables {
		if !utils.IsStringInSlice(table, tablesRename) {
			continue
		}

		tableNew := tablePrefixBackup + table

		if _, err = p.db.ExecContext(ctx, fmt.Sprintf(p.sqlFmtRenameTable, table, tableNew)); err != nil {
			return err
		}

		if p.name == providerPostgres {
			if table == tablePre1U2FDevices || table == tableUserPreferences {
				if _, err = p.db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s RENAME CONSTRAINT %s_pkey TO %s_pkey;`,
					tableNew, table, tableNew)); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigratePre1To1Rollback(ctx context.Context, up bool) (err error) {
	if up {
		migration, err := loadMigration(p.name, 1, false)
		if err != nil {
			return err
		}

		if _, err = p.db.ExecContext(ctx, migration.Query); err != nil {
			return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
		}
	}

	tables, err := p.SchemaTables(ctx)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if !strings.HasPrefix(table, tablePrefixBackup) {
			continue
		}

		tableNew := strings.Replace(table, tablePrefixBackup, "", 1)
		if _, err = p.db.ExecContext(ctx, fmt.Sprintf(p.sqlFmtRenameTable, table, tableNew)); err != nil {
			return err
		}

		if p.name == providerPostgres && (tableNew == tablePre1U2FDevices || tableNew == tableUserPreferences) {
			if _, err = p.db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s RENAME CONSTRAINT %s_pkey TO %s_pkey;`,
				tableNew, table, tableNew)); err != nil {
				continue
			}
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigratePre1To1AuthenticationLogs(ctx context.Context) (err error) {
	for page := 0; true; page++ {
		attempts, err := p.schemaMigratePre1To1AuthenticationLogsGetRows(ctx, page)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}

			return err
		}

		for _, attempt := range attempts {
			_, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1To1InsertAuthenticationLogs), tableAuthenticationLogs), attempt.Username, attempt.Successful, attempt.Time)
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

func (p *SQLProvider) schemaMigratePre1To1AuthenticationLogsGetRows(ctx context.Context, page int) (attempts []model.AuthenticationAttempt, err error) {
	rows, err := p.db.QueryxContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1To1SelectAuthenticationLogs), tablePrefixBackup+tableAuthenticationLogs), page*100)
	if err != nil {
		return nil, err
	}

	attempts = make([]model.AuthenticationAttempt, 0, 100)

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

		attempts = append(attempts, model.AuthenticationAttempt{Username: username, Successful: successful, Time: time.Unix(timestamp, 0)})
	}

	return attempts, nil
}

func (p *SQLProvider) schemaMigratePre1To1TOTP(ctx context.Context) (err error) {
	rows, err := p.db.QueryxContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1SelectTOTPConfigurations), tablePrefixBackup+tablePre1TOTPSecrets))
	if err != nil {
		return err
	}

	var totpConfigs []model.TOTPConfiguration

	defer func() {
		if err := rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	for rows.Next() {
		var username, secret string

		err = rows.Scan(&username, &secret)
		if err != nil {
			return err
		}

		encryptedSecret, err := p.encrypt([]byte(secret))
		if err != nil {
			return err
		}

		totpConfigs = append(totpConfigs, model.TOTPConfiguration{Username: username, Secret: encryptedSecret})
	}

	for _, config := range totpConfigs {
		_, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1To1InsertTOTPConfiguration), tableTOTPConfigurations), config.Username, p.config.TOTP.Issuer, p.config.TOTP.Period, config.Secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigratePre1To1U2F(ctx context.Context) (err error) {
	rows, err := p.db.Queryx(fmt.Sprintf(p.db.Rebind(queryFmtPre1To1SelectU2FDevices), tablePrefixBackup+tablePre1U2FDevices))
	if err != nil {
		return err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	var devices []model.U2FDevice

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

		encryptedPublicKey, err := p.encrypt(publicKey)
		if err != nil {
			return err
		}

		devices = append(devices, model.U2FDevice{Username: username, KeyHandle: keyHandle, PublicKey: encryptedPublicKey})
	}

	for _, device := range devices {
		_, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1To1InsertU2FDevice), tablePre1U2FDevices), device.Username, device.KeyHandle, device.PublicKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate1ToPre1(ctx context.Context) (err error) {
	tables, err := p.SchemaTables(ctx)
	if err != nil {
		return err
	}

	tablesRename := []string{
		tableMigrations,
		tableTOTPConfigurations,
		tableIdentityVerification,
		tablePre1U2FDevices,
		tableDuoDevices,
		tableUserPreferences,
		tableAuthenticationLogs,
		tableEncryption,
	}

	if err = p.schemaMigratePre1Rename(ctx, tables, tablesRename); err != nil {
		return err
	}

	if _, err := p.db.ExecContext(ctx, queryCreatePre1); err != nil {
		return err
	}

	if _, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1InsertUserPreferencesFromSelect),
		tableUserPreferences, tablePrefixBackup+tableUserPreferences)); err != nil {
		return err
	}

	if err = p.schemaMigrate1ToPre1AuthenticationLogs(ctx); err != nil {
		return err
	}

	if err = p.schemaMigrate1ToPre1U2F(ctx); err != nil {
		return err
	}

	if err = p.schemaMigrate1ToPre1TOTP(ctx); err != nil {
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

func (p *SQLProvider) schemaMigrate1ToPre1AuthenticationLogs(ctx context.Context) (err error) {
	for page := 0; true; page++ {
		attempts, err := p.schemaMigrate1ToPre1AuthenticationLogsGetRows(ctx, page)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}

			return err
		}

		for _, attempt := range attempts {
			_, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1InsertAuthenticationLogs), tableAuthenticationLogs), attempt.Username, attempt.Successful, attempt.Time.Unix())
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

func (p *SQLProvider) schemaMigrate1ToPre1AuthenticationLogsGetRows(ctx context.Context, page int) (attempts []model.AuthenticationAttempt, err error) {
	rows, err := p.db.QueryxContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1SelectAuthenticationLogs), tablePrefixBackup+tableAuthenticationLogs), page*100)
	if err != nil {
		return nil, err
	}

	attempts = make([]model.AuthenticationAttempt, 0, 100)

	var attempt model.AuthenticationAttempt
	for rows.Next() {
		err = rows.StructScan(&attempt)
		if err != nil {
			return nil, err
		}

		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

func (p *SQLProvider) schemaMigrate1ToPre1TOTP(ctx context.Context) (err error) {
	rows, err := p.db.QueryxContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmtPre1SelectTOTPConfigurations), tablePrefixBackup+tableTOTPConfigurations))
	if err != nil {
		return err
	}

	var totpConfigs []model.TOTPConfiguration

	defer func() {
		if err := rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	for rows.Next() {
		var (
			username         string
			secretCipherText []byte
		)

		err = rows.Scan(&username, &secretCipherText)
		if err != nil {
			return err
		}

		secretClearText, err := p.decrypt(secretCipherText)
		if err != nil {
			return err
		}

		totpConfigs = append(totpConfigs, model.TOTPConfiguration{Username: username, Secret: secretClearText})
	}

	for _, config := range totpConfigs {
		_, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1InsertTOTPConfiguration), tablePre1TOTPSecrets), config.Username, config.Secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate1ToPre1U2F(ctx context.Context) (err error) {
	rows, err := p.db.QueryxContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1SelectU2FDevices), tablePrefixBackup+tablePre1U2FDevices))
	if err != nil {
		return err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	var (
		devices []model.U2FDevice
		device  model.U2FDevice
	)

	for rows.Next() {
		err = rows.StructScan(&device)
		if err != nil {
			return err
		}

		device.PublicKey, err = p.decrypt(device.PublicKey)
		if err != nil {
			return err
		}

		devices = append(devices, device)
	}

	for _, device := range devices {
		_, err = p.db.ExecContext(ctx, fmt.Sprintf(p.db.Rebind(queryFmt1ToPre1InsertU2FDevice), tablePre1U2FDevices), device.Username, base64.StdEncoding.EncodeToString(device.KeyHandle), base64.StdEncoding.EncodeToString(device.PublicKey))
		if err != nil {
			return err
		}
	}

	return nil
}
