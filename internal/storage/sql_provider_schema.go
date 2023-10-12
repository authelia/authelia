package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

// SchemaTables returns a list of tables from the storage provider.
func (p *SQLProvider) SchemaTables(ctx context.Context) (tables []string, err error) {
	var rows *sqlx.Rows

	switch p.schema {
	case "":
		rows, err = p.db.QueryxContext(ctx, p.sqlSelectExistingTables)
	default:
		rows, err = p.db.QueryxContext(ctx, p.sqlSelectExistingTables, p.schema)
	}

	if err != nil {
		return tables, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
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

// SchemaVersion returns the version of the schema from the storage provider.
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

	var tablesV1 = []string{tableDuoDevices, tableEncryption, tableIdentityVerification, tableMigrations, tableTOTPConfigurations}

	if utils.IsStringSliceContainsAll(tablesPre1, tables) {
		if utils.IsStringSliceContainsAny(tablesV1, tables) {
			return -2, errors.New("pre1 schema contains v1 tables it shouldn't contain")
		}

		return -1, nil
	}

	return 0, nil
}

// SchemaLatestVersion returns the latest version available for migration for the storage provider.
func (p *SQLProvider) SchemaLatestVersion() (version int, err error) {
	return latestMigrationVersion(p.name)
}

// SchemaMigrationHistory returns the storage provider migration history rows.
func (p *SQLProvider) SchemaMigrationHistory(ctx context.Context) (migrations []model.Migration, err error) {
	rows, err := p.db.QueryxContext(ctx, p.sqlSelectMigrations)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	var migration model.Migration

	for rows.Next() {
		err = rows.StructScan(&migration)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// SchemaMigrationsUp returns a list of storage provider up migrations available between the current version
// and the provided version.
func (p *SQLProvider) SchemaMigrationsUp(ctx context.Context, version int) (migrations []model.SchemaMigration, err error) {
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

// SchemaMigrationsDown returns a list of storage provider down migrations available between the current version
// and the provided version.
func (p *SQLProvider) SchemaMigrationsDown(ctx context.Context, version int) (migrations []model.SchemaMigration, err error) {
	current, err := p.SchemaVersion(ctx)
	if err != nil {
		return migrations, err
	}

	if current <= version {
		return migrations, ErrNoAvailableMigrations
	}

	return loadMigrations(p.name, current, version)
}

// SchemaMigrate migrates from the storage provider's current schema version to the provided schema version.
func (p *SQLProvider) SchemaMigrate(ctx context.Context, up bool, version int) (err error) {
	var (
		tx   *sqlx.Tx
		conn SQLXConnection
	)

	if p.name != providerMySQL {
		if tx, err = p.db.BeginTxx(ctx, nil); err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		conn = tx
	} else {
		conn = p.db
	}

	currentVersion, err := p.SchemaVersion(ctx)
	if err != nil {
		return err
	}

	if currentVersion != 0 {
		if err = p.schemaMigrateLock(ctx, conn); err != nil {
			return err
		}
	}

	if err = schemaMigrateChecks(p.name, up, version, currentVersion); err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}

		return err
	}

	if err = p.schemaMigrate(ctx, conn, currentVersion, version); err != nil {
		if tx != nil && err == ErrNoMigrationsFound {
			_ = tx.Rollback()
		}

		return err
	}

	if tx != nil {
		if err = tx.Commit(); err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				return fmt.Errorf("failed to commit the transaction with: commit error: %w, rollback error: %+v", err, rerr)
			}

			return fmt.Errorf("failed to commit the transaction but it has been rolled back: commit error: %w", err)
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate(ctx context.Context, conn SQLXConnection, prior, target int) (err error) {
	migrations, err := loadMigrations(p.name, prior, target)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return ErrNoMigrationsFound
	}

	p.log.Infof(logFmtMigrationFromTo, strconv.Itoa(prior), strconv.Itoa(migrations[len(migrations)-1].After()))

	for i, migration := range migrations {
		if migration.Up && prior == 0 && i == 1 {
			if err = p.schemaMigrateLock(ctx, conn); err != nil {
				return err
			}
		}

		if err = p.schemaMigrateApply(ctx, conn, migration); err != nil {
			return p.schemaMigrateRollback(ctx, conn, prior, migration.After(), err)
		}
	}

	p.log.Infof(logFmtMigrationComplete, strconv.Itoa(prior), strconv.Itoa(migrations[len(migrations)-1].After()))

	return nil
}

func (p *SQLProvider) schemaMigrateLock(ctx context.Context, conn SQLXConnection) (err error) {
	if p.name != providerPostgres {
		return nil
	}

	if _, err = conn.ExecContext(ctx, fmt.Sprintf(queryFmtPostgreSQLLockTable, tableMigrations, "ACCESS EXCLUSIVE")); err != nil {
		return fmt.Errorf("failed to lock tables: %w", err)
	}

	return nil
}

func (p *SQLProvider) schemaMigrateApply(ctx context.Context, conn SQLXConnection, migration model.SchemaMigration) (err error) {
	if migration.NotEmpty() {
		if _, err = conn.ExecContext(ctx, migration.Query); err != nil {
			return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
		}

		if migration.Version == 1 && migration.Up {
			// Add the schema encryption value if upgrading to v1.
			if err = p.setNewEncryptionCheckValue(ctx, conn, &p.keys.encryption); err != nil {
				return err
			}
		}
	}

	if err = p.schemaMigrateFinalize(ctx, conn, migration); err != nil {
		return err
	}

	return nil
}

func (p *SQLProvider) schemaMigrateFinalize(ctx context.Context, conn SQLXConnection, migration model.SchemaMigration) (err error) {
	if migration.Version == 1 && !migration.Up {
		return nil
	}

	if _, err = conn.ExecContext(ctx, p.sqlInsertMigration, time.Now(), migration.Before(), migration.After(), utils.Version()); err != nil {
		return fmt.Errorf("failed inserting migration record: %w", err)
	}

	p.log.Debugf("Storage schema migrated from version %d to %d", migration.Before(), migration.After())

	return nil
}

func (p *SQLProvider) schemaMigrateRollback(ctx context.Context, conn SQLXConnection, prior, after int, merr error) (err error) {
	switch tx := conn.(type) {
	case *sqlx.Tx:
		return p.schemaMigrateRollbackWithTx(ctx, tx, merr)
	default:
		return p.schemaMigrateRollbackWithoutTx(ctx, prior, after, merr)
	}
}

func (p *SQLProvider) schemaMigrateRollbackWithTx(_ context.Context, tx *sqlx.Tx, merr error) (err error) {
	if err = tx.Rollback(); err != nil {
		return fmt.Errorf("error applying rollback %+v. rollback caused by: %w", err, merr)
	}

	return fmt.Errorf("migration rollback complete. rollback caused by: %w", merr)
}

func (p *SQLProvider) schemaMigrateRollbackWithoutTx(ctx context.Context, prior, after int, merr error) (err error) {
	migrations, err := loadMigrations(p.name, after, prior)
	if err != nil {
		return fmt.Errorf("error loading migrations from version %d to version %d for rollback: %+v. rollback caused by: %w", prior, after, err, merr)
	}

	for _, migration := range migrations {
		if err = p.schemaMigrateApply(ctx, p.db, migration); err != nil {
			return fmt.Errorf("error applying migration version %d to version %d for rollback: %+v. rollback caused by: %w", migration.Before(), migration.After(), err, merr)
		}
	}

	return fmt.Errorf("migration rollback complete. rollback caused by: %w", merr)
}

func (p *SQLProvider) schemaLatestMigration(ctx context.Context) (migration *model.Migration, err error) {
	migration = &model.Migration{}

	if err = p.db.QueryRowxContext(ctx, p.sqlSelectLatestMigration).StructScan(migration); err != nil {
		return nil, err
	}

	return migration, nil
}

func schemaMigrateChecks(providerName string, up bool, targetVersion, currentVersion int) (err error) {
	switch {
	case currentVersion == -1:
		return fmt.Errorf(errFmtMigrationPre1, "up from", errFmtMigrationPre1SuggestedVersion)
	case targetVersion == -1:
		return fmt.Errorf(errFmtMigrationPre1, "down to", fmt.Sprintf("you should downgrade to schema version 1 using the current authelia version then use the suggested authelia version to downgrade to pre1: %s", errFmtMigrationPre1SuggestedVersion))
	}

	if targetVersion == currentVersion {
		return fmt.Errorf(ErrFmtMigrateAlreadyOnTargetVersion, targetVersion, currentVersion)
	}

	latest, err := latestMigrationVersion(providerName)
	if err != nil {
		return err
	}

	if currentVersion > latest {
		return fmt.Errorf(errFmtSchemaCurrentGreaterThanLatestKnown, latest)
	}

	if up {
		if targetVersion < currentVersion {
			return fmt.Errorf(ErrFmtMigrateUpTargetLessThanCurrent, targetVersion, currentVersion)
		}

		if targetVersion == SchemaLatest && latest == currentVersion {
			return ErrSchemaAlreadyUpToDate
		}

		if targetVersion != SchemaLatest && latest < targetVersion {
			return fmt.Errorf(ErrFmtMigrateUpTargetGreaterThanLatest, targetVersion, latest)
		}
	} else {
		if targetVersion < 0 {
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
		return na
	default:
		return strconv.Itoa(version)
	}
}
