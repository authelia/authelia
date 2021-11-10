package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/models"
)

// NewSQLProvider generates a generic SQLProvider to be used with other SQL provider NewUp's.
func NewSQLProvider(name, driverName, dataSourceName string) (provider SQLProvider) {
	db, err := sqlx.Open(driverName, dataSourceName)

	provider = SQLProvider{
		name:       name,
		driverName: driverName,
		db:         db,
		log:        logging.Logger(),
		errOpen:    err,

		sqlFmtRenameTable: queryFmtRenameTable,

		sqlUpsertPreferred2FAMethod: fmt.Sprintf(queryFmtUpsertPreferred2FAMethod, tableUserPreferences),
		sqlSelectPreferred2FAMethod: fmt.Sprintf(queryFmtSelectPreferred2FAMethod, tableUserPreferences),

		sqlInsertIdentityVerification:       fmt.Sprintf(queryFmtInsertIdentityVerification, tableIdentityVerification),
		sqlDeleteIdentityVerification:       fmt.Sprintf(queryFmtDeleteIdentityVerification, tableIdentityVerification),
		sqlSelectExistsIdentityVerification: fmt.Sprintf(queryFmtSelectExistsIdentityVerification, tableIdentityVerification),

		sqlUpsertTOTPConfig: fmt.Sprintf(queryFmtUpsertTOTPConfiguration, tableTOTPConfigurations),
		sqlDeleteTOTPConfig: fmt.Sprintf(queryFmtDeleteTOTPConfiguration, tableTOTPConfigurations),
		sqlSelectTOTPConfig: fmt.Sprintf(queryFmtSelectTOTPConfiguration, tableTOTPConfigurations),

		sqlUpsertU2FDevice: fmt.Sprintf(queryFmtUpsertU2FDevice, tableU2FDevices),
		sqlSelectU2FDevice: fmt.Sprintf(queryFmtSelectU2FDevice, tableU2FDevices),

		sqlInsertAuthenticationAttempt:            fmt.Sprintf(queryFmtInsertAuthenticationLogEntry, tableAuthenticationLogs),
		sqlSelectAuthenticationAttemptsByUsername: fmt.Sprintf(queryFmtSelect1FAAuthenticationLogEntryByUsername, tableAuthenticationLogs),

		sqlInsertMigration:       fmt.Sprintf(queryFmtInsertMigration, tableMigrations),
		sqlSelectLatestMigration: fmt.Sprintf(queryFmtSelectLatestMigration, tableMigrations),
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

	sqlFmtRenameTable string

	sqlUpsertPreferred2FAMethod string
	sqlSelectPreferred2FAMethod string

	sqlInsertIdentityVerification       string
	sqlDeleteIdentityVerification       string
	sqlSelectExistsIdentityVerification string

	sqlUpsertTOTPConfig string
	sqlDeleteTOTPConfig string
	sqlSelectTOTPConfig string

	sqlUpsertU2FDevice string
	sqlSelectU2FDevice string

	sqlInsertAuthenticationAttempt            string
	sqlSelectAuthenticationAttemptsByUsername string

	sqlSelectExistingTables string

	sqlSelectLatestMigration string
	sqlInsertMigration       string
}

// StartupCheck implements the provider startup check interface.
func (p *SQLProvider) StartupCheck() (err error) {
	if p.errOpen != nil {
		return p.errOpen
	}

	if p.name == providerPostgres {
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

	current, err := p.SchemaVersion()
	if err != nil {
		return err
	}

	latest, err := p.SchemaLatestVersion()
	if err != nil {
		return err
	}

	if current > latest {
		return fmt.Errorf(errFmtSchemaCurrentGreaterThanLatestKnown, latest)
	}

	return p.schemaMigrate(current, latest)
}

// SavePreferred2FAMethod save the preferred method for 2FA to the database.
func (p *SQLProvider) SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertPreferred2FAMethod, username, method)

	return err
}

// LoadPreferred2FAMethod load the preferred method for 2FA from the database.
func (p *SQLProvider) LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error) {
	err = p.db.GetContext(ctx, &method, p.sqlSelectPreferred2FAMethod, username)

	switch err {
	case sql.ErrNoRows:
		return "", nil
	case nil:
		return method, err
	default:
		return "", err
	}
}

// SaveIdentityVerification save an identity verification record to the database.
func (p *SQLProvider) SaveIdentityVerification(ctx context.Context, verification models.IdentityVerification) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlInsertIdentityVerification, verification.Token)

	return err
}

// RemoveIdentityVerification remove an identity verification record from the database.
func (p *SQLProvider) RemoveIdentityVerification(ctx context.Context, token string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlDeleteIdentityVerification, token)

	return err
}

// FindIdentityVerification checks if an identity verification record is in the database and active.
func (p *SQLProvider) FindIdentityVerification(ctx context.Context, jti string) (found bool, err error) {
	err = p.db.GetContext(ctx, &found, p.sqlSelectExistsIdentityVerification, jti)
	if err != nil {
		return false, err
	}

	return found, nil
}

// SaveTOTPConfiguration save a TOTP config of a given user in the database.
func (p *SQLProvider) SaveTOTPConfiguration(ctx context.Context, config models.TOTPConfiguration) (err error) {
	// TODO: Encrypt config.Secret here.
	_, err = p.db.ExecContext(ctx, p.sqlUpsertTOTPConfig,
		config.Username,
		config.Algorithm,
		config.Digits,
		config.Period,
		config.Secret,
	)

	return err
}

// DeleteTOTPConfiguration delete a TOTP secret from the database given a username.
func (p *SQLProvider) DeleteTOTPConfiguration(ctx context.Context, username string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlDeleteTOTPConfig, username)

	return err
}

// LoadTOTPConfiguration load a TOTP secret given a username from the database.
func (p *SQLProvider) LoadTOTPConfiguration(ctx context.Context, username string) (config *models.TOTPConfiguration, err error) {
	config = &models.TOTPConfiguration{}

	err = p.db.QueryRowxContext(ctx, p.sqlSelectTOTPConfig, username).StructScan(config)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoTOTPSecret
		}

		return nil, err
	}

	// TODO: Decrypt config.Secret here.
	return config, nil
}

// SaveU2FDevice saves a registered U2F device.
func (p *SQLProvider) SaveU2FDevice(ctx context.Context, device models.U2FDevice) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertU2FDevice, device.Username, device.KeyHandle, device.PublicKey)

	return err
}

// LoadU2FDevice loads a U2F device registration for a given username.
func (p *SQLProvider) LoadU2FDevice(ctx context.Context, username string) (device *models.U2FDevice, err error) {
	device = &models.U2FDevice{
		Username: username,
	}

	err = p.db.GetContext(ctx, device, p.sqlSelectU2FDevice, username)
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
	_, err = p.db.ExecContext(ctx, p.sqlInsertAuthenticationAttempt, attempt.Time, attempt.Successful, attempt.Username)
	return err
}

// LoadAuthenticationLogs retrieve the latest failed authentications from the authentication log.
func (p *SQLProvider) LoadAuthenticationLogs(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []models.AuthenticationAttempt, err error) {
	rows, err := p.db.QueryxContext(ctx, p.sqlSelectAuthenticationAttemptsByUsername, fromDate, username, limit, limit*page)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			p.log.Warnf(logFmtErrClosingConn, err)
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

	return attempts, nil
}
