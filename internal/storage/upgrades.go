package storage

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/authelia/authelia/internal/utils"
)

func (p *SQLProvider) upgradeCreateTableStatements(tx *sql.Tx, statements map[string]string, existingTables []string) error {
	for table, statement := range statements {
		if !utils.IsStringInSlice(table, existingTables) {
			_, err := tx.Exec(fmt.Sprintf(statement, table))
			if err != nil {
				return fmt.Errorf("Unable to create table %s: %v", table, err)
			}
		}
	}

	return nil
}

func (p *SQLProvider) upgradeRunMultipleStatements(tx *sql.Tx, statements []string) error {
	for _, statement := range statements {
		_, err := tx.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

// upgradeFinalize sets the schema version and logs a message, as well as any other future finalization tasks.
func (p *SQLProvider) upgradeFinalize(tx *sql.Tx, version int) error {
	_, err := tx.Exec(p.sqlConfigSetValue, "schema", "version", strconv.Itoa(version))
	if err != nil {
		return err
	}

	p.log.Debugf("%s%d", storageSchemaUpgradeMessage, version)

	return nil
}

// upgradeSchemaToVersion001 upgrades the schema to version 1.
func (p *SQLProvider) upgradeSchemaToVersion001(tx *sql.Tx, tables []string) error {
	err := p.upgradeCreateTableStatements(tx, p.sqlUpgradesCreateTableStatements[1], tables)
	if err != nil {
		return err
	}

	if p.name != "mysql" {
		err = p.upgradeRunMultipleStatements(tx, p.sqlUpgradesCreateTableIndexesStatements[1])
		if err != nil {
			return fmt.Errorf("Unable to create index: %v", err)
		}
	}

	err = p.upgradeFinalize(tx, 1)
	if err != nil {
		return err
	}

	return nil
}

// upgradeSchemaToVersion002 upgrades the schema to faux version 2.
func (p *SQLProvider) upgradeSchemaToVersion002(tx *sql.Tx, tables []string) error {
	err := p.upgradeFinalize(tx, 2)
	if err != nil {
		return err
	}

	p.log.Tracef("tables are %v", tables)

	return nil
}
