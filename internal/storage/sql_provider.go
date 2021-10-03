package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewSQLProvider generates a generic SQLProvider to be used with other SQL provider NewUp's.
func NewSQLProvider(name, driverName, dataSourceName string) (provider SQLProvider) {
	db, err := sqlx.Open(driverName, dataSourceName)

	provider = SQLProvider{
		name:       name,
		driverName: driverName,
		db:         db,
		errOpen:    err,

		sqlUpgradesCreateTableStatements:        sqlUpgradeCreateTableStatements,
		sqlUpgradesCreateTableIndexesStatements: sqlUpgradesCreateTableIndexesStatements,

		sqlRenameTable: queryRenameTable,

		sqlSelectPreferred2FAMethodByUsername: fmt.Sprintf(queryFmtSelectPreferred2FAMethodByUsername, tableUserPreferences),
		sqlUpsertPreferred2FAMethod:           fmt.Sprintf(queryFmtUpsertPreferred2FAMethod, tableUserPreferences),

		sqlSelectExistsIdentityVerificationToken: fmt.Sprintf(queryFmtSelectExistsIdentityVerificationToken, tableIdentityVerificationTokens),
		sqlInsertIdentityVerificationToken:       fmt.Sprintf(queryFmtInsertIdentityVerificationToken, tableIdentityVerificationTokens),
		sqlDeleteIdentityVerificationToken:       fmt.Sprintf(queryFmtDeleteIdentityVerificationToken, tableIdentityVerificationTokens),

		sqlSelectTOTPSecretByUsername: fmt.Sprintf(queryFmtSelectTOTPSecretByUsername, tableTOTPSecrets),
		sqlUpsertTOTPSecret:           fmt.Sprintf(queryFmtUpsertTOTPSecret, tableTOTPSecrets),
		sqlDeleteTOTPSecret:           fmt.Sprintf(queryFmtDeleteTOTPSecret, tableTOTPSecrets),

		sqlSelectU2FDeviceByUsername: fmt.Sprintf(queryFmtSelectU2FDeviceByUsername, tableU2FDevices),
		sqlUpsertU2FDevice:           fmt.Sprintf(queryFmtUpsertU2FDevice, tableU2FDevices),

		sqlInsertAuthenticationAttempt:            fmt.Sprintf(queryFmtInsertAuthenticationAttempt, tableAuthenticationLogs),
		sqlSelectAuthenticationAttemptsByUsername: fmt.Sprintf(queryFmtSelectAuthenticationAttemptsByUsername, tableAuthenticationLogs),
	}

	return provider
}

// SQLProvider is a storage provider persisting data in a SQL database.
type SQLProvider struct {
	db         *sqlx.DB
	log        *logrus.Logger
	name       string
	driverName string
	errOpen    error

	sqlUpgradesCreateTableStatements        map[SchemaVersion]map[string]string
	sqlUpgradesCreateTableIndexesStatements map[SchemaVersion][]string

	sqlRenameTable string

	sqlUpsertPreferred2FAMethod           string
	sqlSelectPreferred2FAMethodByUsername string

	sqlInsertIdentityVerificationToken       string
	sqlDeleteIdentityVerificationToken       string
	sqlSelectExistsIdentityVerificationToken string

	sqlUpsertTOTPSecret           string
	sqlDeleteTOTPSecret           string
	sqlSelectTOTPSecretByUsername string

	sqlUpsertU2FDevice           string
	sqlSelectU2FDeviceByUsername string

	sqlInsertAuthenticationAttempt            string
	sqlSelectAuthenticationAttemptsByUsername string

	sqlSelectExistingTables  string
	sqlSelectLatestMigration string

	sqlConfigSetValue string
	sqlConfigGetValue string
}

// StartupCheck implements the provider startup check interface.
func (p *SQLProvider) StartupCheck(logger *logrus.Logger) (err error) {
	if p.errOpen != nil {
		return p.errOpen
	}

	p.log = logger

	if p.name == "postgres" {
		p.rebind()
	}

	// TODO: Decide if this is needed, or if it should be configurable.
	for i := 0; i < 19; i++ {
		err = p.db.Ping()
		if err == nil {
			break
		}

		time.Sleep(time.Millisecond * 500)
	}

	if err != nil {
		return err
	}

	err = p.migrate()
	if err != nil {
		return err
	}

	return nil
}

// SavePreferred2FAMethod save the preferred method for 2FA to the database.
func (p *SQLProvider) SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertPreferred2FAMethod, username, method)

	return err
}

// LoadPreferred2FAMethod load the preferred method for 2FA from the database.
func (p *SQLProvider) LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error) {
	err = p.db.GetContext(ctx, &method, p.sqlSelectPreferred2FAMethodByUsername, username)

	switch err {
	case sql.ErrNoRows:
		return "", nil
	case nil:
		return method, err
	default:
		return "", err
	}
}

// SaveIdentityVerificationToken save an identity verification token in the database.
func (p *SQLProvider) SaveIdentityVerificationToken(ctx context.Context, token string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlInsertIdentityVerificationToken, token)

	return err
}

// RemoveIdentityVerificationToken remove an identity verification token from the database.
func (p *SQLProvider) RemoveIdentityVerificationToken(ctx context.Context, token string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlDeleteIdentityVerificationToken, token)

	return err
}

// FindIdentityVerificationToken look for an identity verification token in the database.
func (p *SQLProvider) FindIdentityVerificationToken(ctx context.Context, token string) (found bool, err error) {
	err = p.db.GetContext(ctx, &found, p.sqlSelectExistsIdentityVerificationToken, token)
	if err != nil {
		return false, err
	}

	return found, nil
}

// SaveTOTPSecret save a TOTP secret of a given user in the database.
func (p *SQLProvider) SaveTOTPSecret(ctx context.Context, username string, secret string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertTOTPSecret, username, secret)

	return err
}

// DeleteTOTPSecret delete a TOTP secret from the database given a username.
func (p *SQLProvider) DeleteTOTPSecret(ctx context.Context, username string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlDeleteTOTPSecret, username)

	return err
}

// LoadTOTPSecret load a TOTP secret given a username from the database.
func (p *SQLProvider) LoadTOTPSecret(ctx context.Context, username string) (secret string, err error) {
	err = p.db.GetContext(ctx, &secret, p.sqlSelectTOTPSecretByUsername, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoTOTPSecret
		}

		return "", err
	}

	return secret, nil
}

// SaveU2FDeviceHandle save a registered U2F device registration blob.
func (p *SQLProvider) SaveU2FDeviceHandle(ctx context.Context, device models.U2FDevice) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertU2FDevice, device.Username, device.KeyHandle, device.PublicKey)

	return err
}

// LoadU2FDeviceHandle load a U2F device registration blob for a given username.
func (p *SQLProvider) LoadU2FDeviceHandle(ctx context.Context, username string) (device *models.U2FDevice, err error) {
	device = &models.U2FDevice{
		Username: username,
	}

	err = p.db.GetContext(ctx, device, p.sqlSelectU2FDeviceByUsername, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoU2FDeviceHandle
		}

		return nil, err
	}

	return device, nil
}

// AppendAuthenticationLog append a mark to the authentication log.
func (p *SQLProvider) AppendAuthenticationLog(ctx context.Context, attempt models.AuthenticationAttempt) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlInsertAuthenticationAttempt, attempt.Username, attempt.Successful, attempt.Time)
	// p.log.Debugf("DEBUG TEMP: AppendAuthenticationLog(username: %s, time: %d, successful: %v, err: %v)", attempt.Username, attempt.Time.Unix(), attempt.Successful, err)

	return err
}

// LoadAuthenticationAttempts retrieve the latest failed authentications from the authentication log.
func (p *SQLProvider) LoadAuthenticationAttempts(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []models.AuthenticationAttempt, err error) {
	rows, err := p.db.QueryxContext(ctx, p.sqlSelectAuthenticationAttemptsByUsername, fromDate.Unix(), username, limit, limit*page)

	// p.log.Debugf("DEBUG TEMP: LoadAuthenticationAttempts(username: %s, fromDate: %d, limit: %d, page: %d, offset: %d, err: %v)", username, fromDate.Unix(), limit, page, limit*page, err)

	if err != nil {
		return nil, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf("Error occurred closing SQL connection: %v", err)
		}
	}()

	attempts = make([]models.AuthenticationAttempt, 0, limit)

	for rows.Next() {
		var attempt models.AuthenticationAttempt

		err = rows.StructScan(&attempt)
		if err != nil {
			return nil, err
		}

		attempts = append(attempts, attempt)
	}

	// p.log.Debugf("DEBUG TEMP: LoadAuthenticationAttempts(username: %s, fromDate: %d, limit: %d, page: %d, offset: %d, attempts: %d)", username, fromDate.Unix(), limit, page, limit*page, len(attempts))

	return attempts, nil
}

func (p *SQLProvider) getSchemaBasicDetails() (version SchemaVersion, tables []string, err error) {
	rows, err := p.db.Query(p.sqlSelectExistingTables)
	if err != nil {
		return version, tables, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf("Error occurred closing SQL connection: %v", err)
		}
	}()

	var table string

	for rows.Next() {
		err := rows.Scan(&table)
		if err != nil {
			return version, tables, err
		}

		tables = append(tables, table)
	}

	if utils.IsStringInSlice(tableConfig, tables) {
		rows, err := p.db.Query(p.sqlConfigGetValue, "schema", "version")
		if err != nil {
			return version, tables, err
		}

		for rows.Next() {
			err := rows.Scan(&version)
			if err != nil {
				return version, tables, err
			}
		}
	}

	return version, tables, nil
}

func (p *SQLProvider) migrate() (err error) {
	p.log.Debug("Storage schema is being checked to verify it is up to date")

	version, tables, err := p.getSchemaBasicDetails()
	if err != nil {
		return err
	}

	if version < storageSchemaCurrentVersion {
		p.log.Debugf("Storage schema is v%d, latest is v%d", version, storageSchemaCurrentVersion)

		tx, err := p.db.Begin()
		if err != nil {
			return err
		}

		switch version {
		case 0:
			err := p.upgradeSchemaToVersion001(tx, tables)
			if err != nil {
				return p.handleUpgradeFailure(tx, 1, err)
			}

			fallthrough
		default:
			err := tx.Commit()
			if err != nil {
				return err
			}

			p.log.Infof("Storage schema upgrade to v%d completed", storageSchemaCurrentVersion)
		}
	} else {
		p.log.Debug("Storage schema is up to date")
	}

	return nil
}

func (p *SQLProvider) handleUpgradeFailure(tx *sql.Tx, version SchemaVersion, handleErr error) (err error) {
	err = tx.Rollback()
	formattedErr := fmt.Errorf("%s%d: %v", storageSchemaUpgradeErrorText, version, handleErr)

	if err != nil {
		return fmt.Errorf("rollback error occurred: %v (inner error %v)", err, formattedErr)
	}

	return formattedErr
}
