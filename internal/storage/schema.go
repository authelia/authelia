package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/utils"
)

func (p *SQLProvider) SchemaTables() (tables []string, err error) {
	rows, err := p.db.Query(p.sqlSelectExistingTables)
	if err != nil {
		return tables, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf("Error occurred closing SQL connection: %v", err)
		}
	}()

	var table string

	for rows.Next() {
		err = rows.Scan(&table)
		if err != nil {
			return []string{}, err
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (p *SQLProvider) SchemaVersion() (version int, err error) {
	tables, err := p.SchemaTables()
	if err != nil {
		return -2, err
	}

	if len(tables) == 0 {
		return 0, nil
	}

	if utils.IsStringInSlice(tableUserPreferences, tables) && utils.IsStringInSlice(tablePre1TOTPSecrets, tables) &&
		utils.IsStringInSlice(tableU2FDevices, tables) && utils.IsStringInSlice(tableAuthenticationLogs, tables) &&
		utils.IsStringInSlice(tablePre1IdentityVerificationTokens, tables) && !utils.IsStringInSlice(tableMigrations, tables) {

		if len(tables) == 5 || len(tables) == 6 && utils.IsStringInSlice(tablePre1Config, tables) {
			return -1, nil
		}

		return -2, errors.New("unknown schema state")
	}

	if utils.IsStringInSlice(tableMigrations, tables) {
		migration, err := p.schemaLatestMigration()
		if err != nil {
			return -2, err
		}

		return migration.Current, nil
	}

	return -2, errors.New("unknown schema state")
}

func (p *SQLProvider) schemaLatestMigration() (migration *models.Migration, err error) {
	migration = &models.Migration{}

	err = p.db.QueryRowx(p.sqlSelectLatestMigration).StructScan(migration)
	if err != nil {
		return nil, err
	}

	return migration, nil
}

func (p *SQLProvider) SchemaMigrateLatest() (err error) {
	currentVersion, err := p.SchemaVersion()
	if err != nil {
		p.log.Fatal(err)
	}

	err = p.schemaMigrate(currentVersion, 2147483647)
	if err != nil {
		return err
	}

	return nil
}

func (p *SQLProvider) schemaMigrate(prior, target int) (err error) {
	migrations, err := loadMigrations(p.name, prior, target)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		p.log.Infof("Storage schema is up to date")

		return nil
	}

	var trackPrior = prior
	if prior == -1 {
		p.log.Infof("Storage schema migration from pre1 to %d is being attempted", migrations[len(migrations)-1].Version)
		err = p.schemaMigratePre1To1()
		if err != nil {
			if errRollback := p.schemaMigratePre1To1Rollback(); errRollback != nil {
				return fmt.Errorf(errFmtFailedMigrationPre1, err)
			}

			return fmt.Errorf(errFmtFailedMigrationPre1, err)
		}

		trackPrior = 1
	} else {
		p.log.Infof("Storage schema migration from %d to %d is being atttempted", prior, migrations[len(migrations)-1].Version)
	}

	for _, migration := range migrations {
		if migration.Version <= trackPrior {
			continue
		}

		err = p.schemaMigrateApply(trackPrior, migration)
		if err != nil {
			return p.schemaMigrateRollback(prior, trackPrior, err)
		}

		trackPrior = migration.Version
	}

	if prior == -1 {
		p.log.Infof("Storage schema migration from pre1 to %d is complete", migrations[len(migrations)-1].Version)
	} else {
		p.log.Infof("Storage schema migration from %d to %d is complete", prior, migrations[len(migrations)-1].Version)
	}

	return nil
}

func (p *SQLProvider) schemaMigrateRollback(prior, trackPrior int, migrateErr error) (err error) {
	migrations, err := loadMigrations(p.name, trackPrior+1, prior)
	if err != nil {
		return fmt.Errorf("error loading down migrations for rollback: %+v. rollback caused by: %+v", err, migrateErr)
	}

	for _, migration := range migrations {
		err = p.schemaMigrateApply(trackPrior, migration)
		if err != nil {
			return fmt.Errorf("error applyinng down migration v%d: %+v. rollback caused by: %+v", migration.Version, err, migrateErr)
		}
	}

	if prior == -1 {

	}

	return fmt.Errorf("migration rollback complete. rollback caused by: %+v", migrateErr)
}

func (p *SQLProvider) schemaMigrateApply(prior int, migration schemaMigration) (err error) {
	_, err = p.db.Exec(migration.Query)
	if err != nil {
		return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
	}

	// Skip the migration history insertion in a migration to v0.
	if migration.Version == 1 && !migration.Up {
		return nil
	}

	return p.schemaMigrateFinalize(prior, migration)
}

func (p *SQLProvider) schemaMigrateFinalize(prior int, migration schemaMigration) (err error) {
	target := migration.Version
	if !migration.Up {
		target = migration.Version - 1
	}

	// TODO: Add Version.
	_, err = p.db.Exec(p.sqlInsertMigration, time.Now(), prior, target, utils.Version())
	if err != nil {
		return err
	}

	p.log.Debugf("Storage schema migrated from version %d to %d", prior, target)

	return nil
}
