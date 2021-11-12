package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/utils"
)

// SchemaTables returns a list of tables.
func (p *SQLProvider) SchemaTables(ctx context.Context) (tables []string, err error) {
	rows, err := p.db.QueryxContext(ctx, p.sqlSelectExistingTables)
	if err != nil {
		return tables, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
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

// SchemaVersion returns the version of the schema.
func (p *SQLProvider) SchemaVersion(ctx context.Context) (version int, err error) {
	tables, err := p.SchemaTables(ctx)
	if err != nil {
		return -2, err
	}

	if len(tables) == 0 {
		return 0, nil
	}

	if utils.IsStringInSlice(tableMigrations, tables) {
		migration, err := p.schemaLatestMigration(ctx)
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

func (p *SQLProvider) schemaLatestMigration(ctx context.Context) (migration *models.Migration, err error) {
	migration = &models.Migration{}

	err = p.db.QueryRowxContext(ctx, p.sqlSelectLatestMigration).StructScan(migration)
	if err != nil {
		return nil, err
	}

	return migration, nil
}

// SchemaEncryptionChangeKey uses the currently configured key to decrypt values in the database and the key provided
// by this command to encrypt the values again and update them using a transaction.
func (p *SQLProvider) SchemaEncryptionChangeKey(ctx context.Context, encryptionKey string) (err error) {
	tx, err := p.db.Beginx()
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(encryptionKey))

	for page := 0; true; page++ {
		configs, err := p.LoadTOTPConfigurations(ctx, page, 10)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("%v: %w", rollbackErr, err)
			}

			return err
		}

		for _, config := range configs {
			if config.Secret, err = utils.Encrypt(config.Secret, &key); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("%v: %w", rollbackErr, err)
				}

				return err
			}

			if err = p.UpdateTOTPConfigurationSecret(ctx, config); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("%v: %w", rollbackErr, err)
				}

				return err
			}
		}

		if len(configs) != 10 {
			break
		}
	}

	return tx.Commit()
}

// SchemaMigrationHistory returns migration history rows.
func (p *SQLProvider) SchemaMigrationHistory(ctx context.Context) (migrations []models.Migration, err error) {
	rows, err := p.db.QueryxContext(ctx, p.sqlSelectMigrations)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	var migration models.Migration

	for rows.Next() {
		err = rows.StructScan(&migration)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// SchemaMigrate migrates from the current version to the provided version.
func (p *SQLProvider) SchemaMigrate(ctx context.Context, up bool, version int) (err error) {
	currentVersion, err := p.SchemaVersion(ctx)
	if err != nil {
		return err
	}

	if err = schemeMigrateChecks(p.name, up, version, currentVersion); err != nil {
		return err
	}

	return p.schemaMigrate(ctx, currentVersion, version)
}

func (p *SQLProvider) schemaMigrate(ctx context.Context, prior, target int) (err error) {
	migrations, err := loadMigrations(p.name, prior, target)
	if err != nil {
		return err
	}

	switch {
	case prior == -1:
		p.log.Infof(logFmtMigrationFromTo, "pre1", strconv.Itoa(migrations[len(migrations)-1].After()))

		err = p.schemaMigratePre1To1(ctx)
		if err != nil {
			if errRollback := p.schemaMigratePre1To1Rollback(ctx, true); errRollback != nil {
				return fmt.Errorf(errFmtFailedMigrationPre1, err)
			}

			return fmt.Errorf(errFmtFailedMigrationPre1, err)
		}
	case target == -1:
		p.log.Infof(logFmtMigrationFromTo, strconv.Itoa(prior), "pre1")
	default:
		p.log.Infof(logFmtMigrationFromTo, strconv.Itoa(prior), strconv.Itoa(migrations[len(migrations)-1].After()))
	}

	for _, migration := range migrations {
		if prior == -1 && migration.Version == 1 {
			// Skip migration version 1 when upgrading from pre1 as it's applied as part of the pre1 upgrade.
			continue
		}

		err = p.schemaMigrateApply(ctx, migration)
		if err != nil {
			return p.schemaMigrateRollback(ctx, prior, migration.After(), err)
		}
	}

	switch {
	case prior == -1:
		p.log.Infof(logFmtMigrationComplete, "pre1", strconv.Itoa(migrations[len(migrations)-1].After()))
	case target == -1:
		err = p.schemaMigrate1ToPre1(ctx)
		if err != nil {
			if errRollback := p.schemaMigratePre1To1Rollback(ctx, false); errRollback != nil {
				return fmt.Errorf(errFmtFailedMigrationPre1, err)
			}

			return fmt.Errorf(errFmtFailedMigrationPre1, err)
		}

		p.log.Infof(logFmtMigrationComplete, strconv.Itoa(prior), "pre1")
	default:
		p.log.Infof(logFmtMigrationComplete, strconv.Itoa(prior), strconv.Itoa(migrations[len(migrations)-1].After()))
	}

	return nil
}

func (p *SQLProvider) schemaMigrateRollback(ctx context.Context, prior, after int, migrateErr error) (err error) {
	migrations, err := loadMigrations(p.name, after, prior)
	if err != nil {
		return fmt.Errorf("error loading migrations from version %d to version %d for rollback: %+v. rollback caused by: %+v", prior, after, err, migrateErr)
	}

	for _, migration := range migrations {
		if prior == -1 && !migration.Up && migration.Version == 1 {
			continue
		}

		err = p.schemaMigrateApply(ctx, migration)
		if err != nil {
			return fmt.Errorf("error applying migration version %d to version %d for rollback: %+v. rollback caused by: %+v", migration.Before(), migration.After(), err, migrateErr)
		}
	}

	if prior == -1 {
		if err = p.schemaMigrate1ToPre1(ctx); err != nil {
			return fmt.Errorf("error applying migration version 1 to version pre1 for rollback: %+v. rollback caused by: %+v", err, migrateErr)
		}
	}

	return fmt.Errorf("migration rollback complete. rollback caused by: %+v", migrateErr)
}

func (p *SQLProvider) schemaMigrateApply(ctx context.Context, migration SchemaMigration) (err error) {
	_, err = p.db.ExecContext(ctx, migration.Query)
	if err != nil {
		return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
	}

	// Skip the migration history insertion in a migration to v0.
	if migration.Version == 1 && !migration.Up {
		return nil
	}

	return p.schemaMigrateFinalize(ctx, migration)
}

func (p SQLProvider) schemaMigrateFinalize(ctx context.Context, migration SchemaMigration) (err error) {
	return p.schemaMigrateFinalizeAdvanced(ctx, migration.Before(), migration.After())
}

func (p *SQLProvider) schemaMigrateFinalizeAdvanced(ctx context.Context, before, after int) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlInsertMigration, time.Now(), before, after, utils.Version())
	if err != nil {
		return err
	}

	p.log.Debugf("Storage schema migrated from version %d to %d", before, after)

	return nil
}

// SchemaMigrationsUp returns a list of migrations up available between the current version and the provided version.
func (p *SQLProvider) SchemaMigrationsUp(ctx context.Context, version int) (migrations []SchemaMigration, err error) {
	current, err := p.SchemaVersion(ctx)
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

// SchemaMigrationsDown returns a list of migrations down available between the current version and the provided version.
func (p *SQLProvider) SchemaMigrationsDown(ctx context.Context, version int) (migrations []SchemaMigration, err error) {
	current, err := p.SchemaVersion(ctx)
	if err != nil {
		return migrations, err
	}

	if current <= version {
		return migrations, ErrNoAvailableMigrations
	}

	return loadMigrations(p.name, current, version)
}

// SchemaLatestVersion returns the latest version available for migration..
func (p *SQLProvider) SchemaLatestVersion() (version int, err error) {
	return latestMigrationVersion(p.name)
}

func schemeMigrateChecks(providerName string, up bool, targetVersion, currentVersion int) (err error) {
	if targetVersion == currentVersion {
		return fmt.Errorf(ErrFmtMigrateAlreadyOnTargetVersion, targetVersion, currentVersion)
	}

	if up {
		if targetVersion < currentVersion {
			return fmt.Errorf(ErrFmtMigrateUpTargetLessThanCurrent, targetVersion, currentVersion)
		}

		latest, err := latestMigrationVersion(providerName)
		if err != nil {
			return err
		}

		if targetVersion == SchemaLatest && latest == currentVersion {
			return ErrSchemaAlreadyUpToDate
		}

		if targetVersion != SchemaLatest && latest < targetVersion {
			return fmt.Errorf(ErrFmtMigrateUpTargetGreaterThanLatest, targetVersion, latest)
		}
	} else {
		if targetVersion < -1 {
			return fmt.Errorf(ErrFmtMigrateDownTargetLessThanMinimum, targetVersion)
		}

		if targetVersion > currentVersion {
			return fmt.Errorf(ErrFmtMigrateDownTargetGreaterThanCurrent, targetVersion, currentVersion)
		}
	}

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
