package storage

import (
	"fmt"
	"strconv"
	"time"

	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/utils"
)

func (p *SQLProvider) SchemaMigrationsUp(version int) (migrations []SchemaMigration, err error) {
	current, err := p.SchemaVersion()
	if err != nil {
		return migrations, err
	}

	if version == 0 {
		version = SchemaLatest
	}

	if current >= version {
		return migrations, ErrNoAvailableMigrations
	}

	return loadMigrations(p.name, current, version)
}

func (p *SQLProvider) SchemaMigrationsDown(version int) (migrations []SchemaMigration, err error) {
	current, err := p.SchemaVersion()
	if err != nil {
		return migrations, err
	}

	return loadMigrations(p.name, current, version)
}

// SchemaTables returns a list of tables.
func (p *SQLProvider) SchemaTables() (tables []string, err error) {
	rows, err := p.db.Query(p.sqlSelectExistingTables)
	if err != nil {
		return tables, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf(logFmtErrClosingConn, err)
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

// SchemaLatestVersion returns the latest version available for migration..
func (p *SQLProvider) SchemaLatestVersion() (version int, err error) {
	migrations, err := loadMigrations(p.name, 0, SchemaLatest)
	if err != nil {
		return 0, err
	}

	return migrations[len(migrations)-1].Version, nil
}

// SchemaVersion returns the version of the schema.
func (p *SQLProvider) SchemaVersion() (version int, err error) {
	tables, err := p.SchemaTables()
	if err != nil {
		return -2, err
	}

	if len(tables) == 0 {
		return 0, nil
	}

	if utils.IsStringInSlice(tableMigrations, tables) {
		migration, err := p.schemaLatestMigration()
		if err != nil {
			return -2, err
		}

		return migration.After, nil
	}

	if utils.IsStringInSlice(tableUserPreferences, tables) && utils.IsStringInSlice(tablePre1TOTPSecrets, tables) &&
		utils.IsStringInSlice(tableU2FDevices, tables) && utils.IsStringInSlice(tableAuthenticationLogs, tables) &&
		utils.IsStringInSlice(tablePre1IdentityVerificationTokens, tables) && !utils.IsStringInSlice(tableMigrations, tables) {
		return -1, nil
	}

	// TODO: Decide if we want to support external tables.
	// return -2, ErrUnknownSchemaState
	return 0, nil
}

// SchemaMigrate migrates from the current version to the provided version.
func (p *SQLProvider) SchemaMigrate(version int) (err error) {
	currentVersion, err := p.SchemaVersion()
	if err != nil {
		return err
	}

	return p.schemaMigrate(currentVersion, version)
}

func (p *SQLProvider) schemaLatestMigration() (migration *models.Migration, err error) {
	migration = &models.Migration{}

	err = p.db.QueryRowx(p.sqlSelectLatestMigration).StructScan(migration)
	if err != nil {
		return nil, err
	}

	return migration, nil
}

func (p *SQLProvider) schemaMigrate(prior, target int) (err error) {
	up := prior < target

	migrations, err := loadMigrations(p.name, prior, target)
	if err != nil {
		return err
	}

	if len(migrations) == 0 && (target != -1 || prior <= 0) {
		p.log.Infof("Storage schema is up to date")

		return nil
	}

	var trackPrior = prior

	if prior == -1 {
		p.log.Infof(logFmtMigrationFromTo, "pre1", strconv.Itoa(migrations[len(migrations)-1].After()))

		err = p.schemaMigratePre1To1()
		if err != nil {
			if errRollback := p.schemaMigratePre1To1Rollback(true); errRollback != nil {
				return fmt.Errorf(errFmtFailedMigrationPre1, err)
			}

			return fmt.Errorf(errFmtFailedMigrationPre1, err)
		}

		trackPrior = 1
	} else if target == -1 {
		p.log.Infof(logFmtMigrationFromTo, strconv.Itoa(prior), "pre1")
	} else {
		p.log.Infof(logFmtMigrationFromTo, strconv.Itoa(prior), strconv.Itoa(migrations[len(migrations)-1].After()))
	}

	for _, migration := range migrations {
		if target == -1 && migration.Version == 1 {
			continue
		}

		// Skip same version number migrations.
		if up && migration.Version <= trackPrior {
			continue
		}

		// Skip same version number migrations.
		if !up && migration.Version > trackPrior {
			continue
		}

		err = p.schemaMigrateApply(trackPrior, migration)
		if err != nil {
			return p.schemaMigrateRollback(prior, trackPrior, err)
		}

		trackPrior = migration.Version
	}

	if prior == -1 {
		p.log.Infof(logFmtMigrationComplete, "pre1", strconv.Itoa(migrations[len(migrations)-1].After()))
	} else if target == -1 {
		err = p.schemaMigrate1ToPre1()
		if err != nil {
			if errRollback := p.schemaMigratePre1To1Rollback(false); errRollback != nil {
				return fmt.Errorf(errFmtFailedMigrationPre1, err)
			}

			return fmt.Errorf(errFmtFailedMigrationPre1, err)
		}
		p.log.Infof(logFmtMigrationComplete, strconv.Itoa(prior), "pre1")
	} else {
		p.log.Infof(logFmtMigrationComplete, strconv.Itoa(prior), strconv.Itoa(migrations[len(migrations)-1].After()))
	}

	return nil
}

func (p *SQLProvider) schemaMigrateRollback(prior, trackPrior int, migrateErr error) (err error) {
	migrations, err := loadMigrations(p.name, trackPrior+1, prior)
	if err != nil {
		return fmt.Errorf("error loading migrations for rollback: %+v. rollback caused by: %+v", err, migrateErr)
	}

	for _, migration := range migrations {
		err = p.schemaMigrateApply(trackPrior, migration)
		if err != nil {
			return fmt.Errorf("error applying migration v%d for rollback: %+v. rollback caused by: %+v", migration.Version, err, migrateErr)
		}
	}

	if prior == -1 {
		// TODO: implement rollback here.
	}

	return fmt.Errorf("migration rollback complete. rollback caused by: %+v", migrateErr)
}

func (p *SQLProvider) schemaMigrateApply(prior int, migration SchemaMigration) (err error) {
	_, err = p.db.Exec(migration.Query)
	if err != nil {
		return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
	}

	// Skip the migration history insertion in a migration to v0.
	if migration.Version == 1 && !migration.Up {
		return nil
	}

	return p.schemaMigrateFinalize(migration)
}

func (p SQLProvider) schemaMigrateFinalize(migration SchemaMigration) (err error) {
	return p.schemaMigrateFinalizeAdvanced(migration.Before(), migration.After())
}

func (p *SQLProvider) schemaMigrateFinalizeAdvanced(before, after int) (err error) {
	_, err = p.db.Exec(p.sqlInsertMigration, time.Now(), before, after, utils.Version())
	if err != nil {
		return err
	}

	p.log.Debugf("Storage schema migrated from version %d to %d", before, after)

	return nil
}

// SchemaVersionToString returns a version string given a version number.
func SchemaVersionToString(version int) (versionStr string) {
	switch version {
	case -2:
		return "unknown"
	case -1:
		return "pre1"
	case 0:
		return "N/A"
	default:
		return strconv.Itoa(version)
	}
}
