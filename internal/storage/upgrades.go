package storage

import (
	"database/sql"
	"fmt"

	"github.com/authelia/authelia/internal/utils"
)

func (p *SQLProvider) upgradeSchemaVersionTo001(tx *sql.Tx, tables []string) error {
	_, err := tx.Exec(SQLCreateConfigTable)
	if err != nil {
		return err
	}

	_, err = tx.Exec(p.sqlConfigSetValue, "schema", "version", "1")
	if err != nil {
		return err
	}

	if !utils.IsStringInSlice(preferencesTableName, tables) {
		_, err := tx.Exec(p.sqlCreateUserPreferencesTable)
		if err != nil {
			return fmt.Errorf("Unable to create table %s: %v", preferencesTableName, err)
		}
	}

	if !utils.IsStringInSlice(identityVerificationTokensTableName, tables) {
		_, err := tx.Exec(p.sqlCreateIdentityVerificationTokensTable)
		if err != nil {
			return fmt.Errorf("Unable to create table %s: %v", identityVerificationTokensTableName, err)
		}
	}

	if !utils.IsStringInSlice(totpSecretsTableName, tables) {
		_, err := tx.Exec(p.sqlCreateTOTPSecretsTable)
		if err != nil {
			return fmt.Errorf("Unable to create table %s: %v", totpSecretsTableName, err)
		}
	}

	if !utils.IsStringInSlice(u2fDeviceHandlesTableName, tables) {
		_, err := tx.Exec(p.sqlCreateU2FDeviceHandlesTable)
		if err != nil {
			return fmt.Errorf("Unable to create table %s: %v", u2fDeviceHandlesTableName, err)
		}
	}

	if !utils.IsStringInSlice(authenticationLogsTableName, tables) {
		_, err := tx.Exec(p.sqlCreateAuthenticationLogsTable)
		if err != nil {
			return fmt.Errorf("Unable to create table %s: %v", authenticationLogsTableName, err)
		}
	}

	if p.sqlCreateAuthenticationLogsUserTimeIndex != "" {
		_, err = tx.Exec(p.sqlCreateAuthenticationLogsUserTimeIndex)
		if err != nil {
			return fmt.Errorf("Unable to create index on %s: %v", authenticationLogsTableName, err)
		}
	}

	p.log.Debugf("%s %d", storageSchemaUpgradeMessage, 1)

	return nil
}

func (p *SQLProvider) upgradeSchemaVersionTo002(tx *sql.Tx) error {
	_, err := tx.Exec(p.sqlConfigSetValue, "schema", "version", "2")
	if err != nil {
		return err
	}

	p.log.Debugf("%s %d", storageSchemaUpgradeMessage, 2)

	return nil
}
