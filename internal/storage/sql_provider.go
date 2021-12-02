package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/models"
)

// NewSQLProvider generates a generic SQLProvider to be used with other SQL provider NewUp's.
func NewSQLProvider(config *schema.Configuration, name, driverName, dataSourceName string) (provider SQLProvider) {
	db, err := sqlx.Open(driverName, dataSourceName)

	provider = SQLProvider{
		db:         db,
		key:        sha256.Sum256([]byte(config.Storage.EncryptionKey)),
		name:       name,
		driverName: driverName,
		config:     config,
		errOpen:    err,
		log:        logging.Logger(),

		sqlInsertAuthenticationAttempt:            fmt.Sprintf(queryFmtInsertAuthenticationLogEntry, tableAuthenticationLogs),
		sqlSelectAuthenticationAttemptsByUsername: fmt.Sprintf(queryFmtSelect1FAAuthenticationLogEntryByUsername, tableAuthenticationLogs),

		sqlInsertIdentityVerification:       fmt.Sprintf(queryFmtInsertIdentityVerification, tableIdentityVerification),
		sqlDeleteIdentityVerification:       fmt.Sprintf(queryFmtDeleteIdentityVerification, tableIdentityVerification),
		sqlSelectExistsIdentityVerification: fmt.Sprintf(queryFmtSelectExistsIdentityVerification, tableIdentityVerification),

		sqlUpsertTOTPConfig:  fmt.Sprintf(queryFmtUpsertTOTPConfiguration, tableTOTPConfigurations),
		sqlDeleteTOTPConfig:  fmt.Sprintf(queryFmtDeleteTOTPConfiguration, tableTOTPConfigurations),
		sqlSelectTOTPConfig:  fmt.Sprintf(queryFmtSelectTOTPConfiguration, tableTOTPConfigurations),
		sqlSelectTOTPConfigs: fmt.Sprintf(queryFmtSelectTOTPConfigurations, tableTOTPConfigurations),

		sqlUpdateTOTPConfigSecret:           fmt.Sprintf(queryFmtUpdateTOTPConfigurationSecret, tableTOTPConfigurations),
		sqlUpdateTOTPConfigSecretByUsername: fmt.Sprintf(queryFmtUpdateTOTPConfigurationSecretByUsername, tableTOTPConfigurations),

		sqlUpsertU2FDevice: fmt.Sprintf(queryFmtUpsertU2FDevice, tableU2FDevices),
		sqlSelectU2FDevice: fmt.Sprintf(queryFmtSelectU2FDevice, tableU2FDevices),

		sqlUpsertDuoDevice: fmt.Sprintf(queryFmtUpsertDuoDevice, tableDuoDevices),
		sqlDeleteDuoDevice: fmt.Sprintf(queryFmtDeleteDuoDevice, tableDuoDevices),
		sqlSelectDuoDevice: fmt.Sprintf(queryFmtSelectDuoDevice, tableDuoDevices),

		sqlUpsertPreferred2FAMethod: fmt.Sprintf(queryFmtUpsertPreferred2FAMethod, tableUserPreferences),
		sqlSelectPreferred2FAMethod: fmt.Sprintf(queryFmtSelectPreferred2FAMethod, tableUserPreferences),
		sqlSelectUserInfo:           fmt.Sprintf(queryFmtSelectUserInfo, tableTOTPConfigurations, tableU2FDevices, tableDuoDevices, tableUserPreferences),

		sqlInsertMigration:       fmt.Sprintf(queryFmtInsertMigration, tableMigrations),
		sqlSelectMigrations:      fmt.Sprintf(queryFmtSelectMigrations, tableMigrations),
		sqlSelectLatestMigration: fmt.Sprintf(queryFmtSelectLatestMigration, tableMigrations),

		sqlUpsertEncryptionValue: fmt.Sprintf(queryFmtUpsertEncryptionValue, tableEncryption),
		sqlSelectEncryptionValue: fmt.Sprintf(queryFmtSelectEncryptionValue, tableEncryption),

		sqlFmtRenameTable: queryFmtRenameTable,
	}

	return provider
}

// SQLProvider is a storage provider persisting data in a SQL database.
type SQLProvider struct {
	db         *sqlx.DB
	key        [32]byte
	name       string
	driverName string
	config     *schema.Configuration
	errOpen    error

	log *logrus.Logger

	// Table: authentication_logs.
	sqlInsertAuthenticationAttempt            string
	sqlSelectAuthenticationAttemptsByUsername string

	// Table: identity_verification.
	sqlInsertIdentityVerification       string
	sqlDeleteIdentityVerification       string
	sqlSelectExistsIdentityVerification string

	// Table: totp_configurations.
	sqlUpsertTOTPConfig  string
	sqlDeleteTOTPConfig  string
	sqlSelectTOTPConfig  string
	sqlSelectTOTPConfigs string

	sqlUpdateTOTPConfigSecret           string
	sqlUpdateTOTPConfigSecretByUsername string

	// Table: u2f_devices.
	sqlUpsertU2FDevice string
	sqlSelectU2FDevice string

	// Table: duo_devices
	sqlUpsertDuoDevice string
	sqlDeleteDuoDevice string
	sqlSelectDuoDevice string

	// Table: user_preferences.
	sqlUpsertPreferred2FAMethod string
	sqlSelectPreferred2FAMethod string
	sqlSelectUserInfo           string

	// Table: migrations.
	sqlInsertMigration       string
	sqlSelectMigrations      string
	sqlSelectLatestMigration string

	// Table: encryption.
	sqlUpsertEncryptionValue string
	sqlSelectEncryptionValue string

	// Utility.
	sqlSelectExistingTables string
	sqlFmtRenameTable       string
}

// Close the underlying database connection.
func (p *SQLProvider) Close() (err error) {
	return p.db.Close()
}

// StartupCheck implements the provider startup check interface.
func (p *SQLProvider) StartupCheck() (err error) {
	if p.errOpen != nil {
		return fmt.Errorf("error opening database: %w", p.errOpen)
	}

	// TODO: Decide if this is needed, or if it should be configurable.
	for i := 0; i < 19; i++ {
		if err = p.db.Ping(); err == nil {
			break
		}

		time.Sleep(time.Millisecond * 500)
	}

	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	p.log.Infof("Storage schema is being checked for updates")

	ctx := context.Background()

	if err = p.SchemaEncryptionCheckKey(ctx, false); err != nil && !errors.Is(err, ErrSchemaEncryptionVersionUnsupported) {
		return err
	}

	err = p.SchemaMigrate(ctx, true, SchemaLatest)

	switch err {
	case ErrSchemaAlreadyUpToDate:
		p.log.Infof("Storage schema is already up to date")
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("error during schema migrate: %w", err)
	}
}

// SavePreferred2FAMethod save the preferred method for 2FA to the database.
func (p *SQLProvider) SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertPreferred2FAMethod, username, method)

	return err
}

// LoadPreferred2FAMethod load the preferred method for 2FA from the database.
func (p *SQLProvider) LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error) {
	err = p.db.GetContext(ctx, &method, p.sqlSelectPreferred2FAMethod, username)

	switch {
	case err == nil:
		return method, nil
	case errors.Is(err, sql.ErrNoRows):
		return "", nil
	default:
		return "", fmt.Errorf("error selecting preferred two factor method for user '%s': %w", username, err)
	}
}

// LoadUserInfo loads the models.UserInfo from the database.
func (p *SQLProvider) LoadUserInfo(ctx context.Context, username string) (info models.UserInfo, err error) {
	err = p.db.GetContext(ctx, &info, p.sqlSelectUserInfo, username, username, username, username)

	switch {
	case err == nil:
		return info, nil
	case errors.Is(err, sql.ErrNoRows):
		if _, err = p.db.ExecContext(ctx, p.sqlUpsertPreferred2FAMethod, username, authentication.PossibleMethods[0]); err != nil {
			return models.UserInfo{}, fmt.Errorf("error upserting preferred two factor method while selecting user info for user '%s': %w", username, err)
		}

		if err = p.db.GetContext(ctx, &info, p.sqlSelectUserInfo, username, username, username, username); err != nil {
			return models.UserInfo{}, fmt.Errorf("error selecting user info for user '%s': %w", username, err)
		}

		return info, nil
	default:
		return models.UserInfo{}, fmt.Errorf("error selecting user info for user '%s': %w", username, err)
	}
}

// SaveIdentityVerification save an identity verification record to the database.
func (p *SQLProvider) SaveIdentityVerification(ctx context.Context, verification models.IdentityVerification) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertIdentityVerification,
		verification.JTI, verification.IssuedAt, verification.ExpiresAt,
		verification.Username, verification.Action); err != nil {
		return fmt.Errorf("error inserting identity verification: %w", err)
	}

	return nil
}

// RemoveIdentityVerification remove an identity verification record from the database.
func (p *SQLProvider) RemoveIdentityVerification(ctx context.Context, jti string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteIdentityVerification, jti); err != nil {
		return fmt.Errorf("error updating identity verification: %w", err)
	}

	return nil
}

// FindIdentityVerification checks if an identity verification record is in the database and active.
func (p *SQLProvider) FindIdentityVerification(ctx context.Context, jti string) (found bool, err error) {
	if err = p.db.GetContext(ctx, &found, p.sqlSelectExistsIdentityVerification, jti); err != nil {
		return false, fmt.Errorf("error selecting identity verification exists: %w", err)
	}

	return found, nil
}

// SaveTOTPConfiguration save a TOTP configuration of a given user in the database.
func (p *SQLProvider) SaveTOTPConfiguration(ctx context.Context, config models.TOTPConfiguration) (err error) {
	if config.Secret, err = p.encrypt(config.Secret); err != nil {
		return fmt.Errorf("error encrypting the TOTP configuration secret: %v", err)
	}

	if _, err = p.db.ExecContext(ctx, p.sqlUpsertTOTPConfig,
		config.Username, config.Issuer, config.Algorithm, config.Digits, config.Period, config.Secret); err != nil {
		return fmt.Errorf("error upserting TOTP configuration: %w", err)
	}

	return nil
}

// DeleteTOTPConfiguration delete a TOTP configuration from the database given a username.
func (p *SQLProvider) DeleteTOTPConfiguration(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteTOTPConfig, username); err != nil {
		return fmt.Errorf("error deleting TOTP configuration: %w", err)
	}

	return nil
}

// LoadTOTPConfiguration load a TOTP configuration given a username from the database.
func (p *SQLProvider) LoadTOTPConfiguration(ctx context.Context, username string) (config *models.TOTPConfiguration, err error) {
	config = &models.TOTPConfiguration{}

	if err = p.db.QueryRowxContext(ctx, p.sqlSelectTOTPConfig, username).StructScan(config); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoTOTPConfiguration
		}

		return nil, fmt.Errorf("error selecting TOTP configuration: %w", err)
	}

	if config.Secret, err = p.decrypt(config.Secret); err != nil {
		return nil, fmt.Errorf("error decrypting the TOTP secret: %v", err)
	}

	return config, nil
}

// LoadTOTPConfigurations load a set of TOTP configurations.
func (p *SQLProvider) LoadTOTPConfigurations(ctx context.Context, limit, page int) (configs []models.TOTPConfiguration, err error) {
	rows, err := p.db.QueryxContext(ctx, p.sqlSelectTOTPConfigs, limit, limit*page)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return configs, nil
		}

		return nil, fmt.Errorf("error selecting TOTP configurations: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	configs = make([]models.TOTPConfiguration, 0, limit)

	var config models.TOTPConfiguration

	for rows.Next() {
		if err = rows.StructScan(&config); err != nil {
			return nil, fmt.Errorf("error scanning TOTP configuration to struct: %w", err)
		}

		if config.Secret, err = p.decrypt(config.Secret); err != nil {
			return nil, fmt.Errorf("error decrypting the TOTP secret: %v", err)
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// UpdateTOTPConfigurationSecret updates a TOTP configuration secret.
func (p *SQLProvider) UpdateTOTPConfigurationSecret(ctx context.Context, config models.TOTPConfiguration) (err error) {
	switch config.ID {
	case 0:
		_, err = p.db.ExecContext(ctx, p.sqlUpdateTOTPConfigSecretByUsername, config.Secret, config.Username)
	default:
		_, err = p.db.ExecContext(ctx, p.sqlUpdateTOTPConfigSecret, config.Secret, config.ID)
	}

	if err != nil {
		return fmt.Errorf("error updating TOTP configuration secret: %w", err)
	}

	return nil
}

// SaveU2FDevice saves a registered U2F device.
func (p *SQLProvider) SaveU2FDevice(ctx context.Context, device models.U2FDevice) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertU2FDevice, device.Username, device.KeyHandle, device.PublicKey); err != nil {
		return fmt.Errorf("error upserting U2F device secret: %v", err)
	}

	return nil
}

// LoadU2FDevice loads a U2F device registration for a given username.
func (p *SQLProvider) LoadU2FDevice(ctx context.Context, username string) (device *models.U2FDevice, err error) {
	device = &models.U2FDevice{
		Username: username,
	}

	if err = p.db.GetContext(ctx, device, p.sqlSelectU2FDevice, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoU2FDeviceHandle
		}

		return nil, fmt.Errorf("error selecting U2F device: %w", err)
	}

	return device, nil
}

// SavePreferredDuoDevice saves a Duo device.
func (p *SQLProvider) SavePreferredDuoDevice(ctx context.Context, device models.DuoDevice) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertDuoDevice, device.Username, device.Device, device.Method)
	return err
}

// DeletePreferredDuoDevice deletes a Duo device of a given user.
func (p *SQLProvider) DeletePreferredDuoDevice(ctx context.Context, username string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlDeleteDuoDevice, username)
	return err
}

// LoadPreferredDuoDevice loads a Duo device of a given user.
func (p *SQLProvider) LoadPreferredDuoDevice(ctx context.Context, username string) (device *models.DuoDevice, err error) {
	device = &models.DuoDevice{}

	if err := p.db.QueryRowxContext(ctx, p.sqlSelectDuoDevice, username).StructScan(device); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoDuoDevice
		}

		return nil, err
	}

	return device, nil
}

// AppendAuthenticationLog append a mark to the authentication log.
func (p *SQLProvider) AppendAuthenticationLog(ctx context.Context, attempt models.AuthenticationAttempt) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertAuthenticationAttempt,
		attempt.Time, attempt.Successful, attempt.Banned, attempt.Username,
		attempt.Type, attempt.RemoteIP, attempt.RequestURI, attempt.RequestMethod); err != nil {
		return fmt.Errorf("error inserting authentication attempt: %w", err)
	}

	return nil
}

// LoadAuthenticationLogs retrieve the latest failed authentications from the authentication log.
func (p *SQLProvider) LoadAuthenticationLogs(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []models.AuthenticationAttempt, err error) {
	rows, err := p.db.QueryxContext(ctx, p.sqlSelectAuthenticationAttemptsByUsername, fromDate, username, limit, limit*page)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoAuthenticationLogs
		}

		return nil, fmt.Errorf("error selecting authentication logs: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			p.log.Errorf(logFmtErrClosingConn, err)
		}
	}()

	var attempt models.AuthenticationAttempt

	attempts = make([]models.AuthenticationAttempt, 0, limit)

	for rows.Next() {
		if err = rows.StructScan(&attempt); err != nil {
			return nil, err
		}

		attempts = append(attempts, attempt)
	}

	return attempts, nil
}
