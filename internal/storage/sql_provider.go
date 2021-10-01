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
	"github.com/authelia/authelia/v4/internal/utils"
)

// SQLProvider is a storage provider persisting data in a SQL database.
type SQLProvider struct {
	db   *sqlx.DB
	log  *logrus.Logger
	name string

	sqlUpgradesCreateTableStatements        map[SchemaVersion]map[string]string
	sqlUpgradesCreateTableIndexesStatements map[SchemaVersion][]string

	sqlGetPreferencesByUsername     string
	sqlUpsertSecondFactorPreference string

	sqlTestIdentityVerificationTokenExistence string
	sqlInsertIdentityVerificationToken        string
	sqlDeleteIdentityVerificationToken        string

	sqlGetTOTPSecretByUsername string
	sqlUpsertTOTPSecret        string
	sqlDeleteTOTPSecret        string

	sqlGetU2FDeviceHandleByUsername string
	sqlUpsertU2FDeviceHandle        string

	sqlInsertAuthenticationLog         string
	sqlGetFailedAuthenticationAttempts string

	sqlGetExistingTables string

	sqlConfigSetValue string
	sqlConfigGetValue string
}

// LoadPreferred2FAMethod load the preferred method for 2FA from the database.
func (p *SQLProvider) LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error) {
	rows, err := p.db.QueryContext(ctx, p.sqlGetPreferencesByUsername, username)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", nil
	}

	err = rows.Scan(&method)

	return method, err
}

// SavePreferred2FAMethod save the preferred method for 2FA to the database.
func (p *SQLProvider) SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertSecondFactorPreference, username, method)

	return err
}

// FindIdentityVerificationToken look for an identity verification token in the database.
func (p *SQLProvider) FindIdentityVerificationToken(ctx context.Context, token string) (found bool, err error) {
	err = p.db.QueryRowContext(ctx, p.sqlTestIdentityVerificationTokenExistence, token).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
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

// SaveTOTPSecret save a TOTP secret of a given user in the database.
func (p *SQLProvider) SaveTOTPSecret(ctx context.Context, username string, secret string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertTOTPSecret, username, secret)

	return err
}

// LoadTOTPSecret load a TOTP secret given a username from the database.
func (p *SQLProvider) LoadTOTPSecret(ctx context.Context, username string) (secret string, err error) {
	if err := p.db.QueryRowContext(ctx, p.sqlGetTOTPSecretByUsername, username).Scan(&secret); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoTOTPSecret
		}

		return "", err
	}

	return secret, nil
}

// DeleteTOTPSecret delete a TOTP secret from the database given a username.
func (p *SQLProvider) DeleteTOTPSecret(ctx context.Context, username string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlDeleteTOTPSecret, username)

	return err
}

// SaveU2FDeviceHandle save a registered U2F device registration blob.
func (p *SQLProvider) SaveU2FDeviceHandle(ctx context.Context, device models.U2FDevice) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlUpsertU2FDeviceHandle, device.Username, device.KeyHandle, device.PublicKey)

	return err
}

// LoadU2FDeviceHandle load a U2F device registration blob for a given username.
func (p *SQLProvider) LoadU2FDeviceHandle(ctx context.Context, username string) (device *models.U2FDevice, err error) {
	device = &models.U2FDevice{
		Username: username,
	}

	err = p.db.GetContext(ctx, device, p.sqlGetU2FDeviceHandleByUsername, username)
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
	p.log.Debugf("DBG: Log Attempt for %s, %v, %d", attempt.Username, attempt.Successful, attempt.Time.Unix())
	_, err = p.db.ExecContext(ctx, p.sqlInsertAuthenticationLog, attempt.Username, attempt.Successful, attempt.Time)
	if err != nil {
		p.log.Debugf("DBG: Log Attempt Err for %s: %v", attempt.Username, err)
	}

	return err
}

// LoadFailedAuthenticationAttempts retrieve the latest failed authentications from the authentication log.
func (p *SQLProvider) LoadFailedAuthenticationAttempts(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []models.AuthenticationAttempt, err error) {
	p.log.Debugf("DBG: Load Logs for %s from %d", username, fromDate.Unix())
	rows, err := p.db.QueryxContext(ctx, p.sqlGetFailedAuthenticationAttempts, fromDate.Unix(), username, limit, limit*page)
	if err != nil {
		p.log.Debugf("DBG: Err %s, %v", username, err)

		return nil, err
	}

	attempts = make([]models.AuthenticationAttempt, 0, limit)

	p.log.Debugf("DBG: Reading rows %s", username)
	for rows.Next() {
		var attempt models.AuthenticationAttempt

		err = rows.StructScan(&attempt)
		if err != nil {
			closeErr := rows.Close()
			if closeErr != nil {
				p.log.Debugf("DBG: Err scan/close %s: %v / %v", username, err, closeErr)
				return nil, fmt.Errorf("%w, error occured closing connection: %+v", err, closeErr)
			}

			p.log.Debugf("DBG: Err scan %s: %v", username, err)
			return nil, err
		}

		p.log.Debugf("DBG: attempt row loaded for %s: %v %v", username, attempt.Successful, attempt.Time.Time)
		attempts = append(attempts, attempt)
	}

	p.log.Debugf("DBG: attempts returned for %s: %d", username, len(attempts))

	return attempts, nil
}

func (p *SQLProvider) initialize(db *sqlx.DB) (err error) {
	p.db = db
	p.log = logging.Logger()

	return p.upgrade()
}

func (p *SQLProvider) getSchemaBasicDetails() (version SchemaVersion, tables []string, err error) {
	rows, err := p.db.Query(p.sqlGetExistingTables)
	if err != nil {
		return version, tables, err
	}

	defer rows.Close()

	var table string

	for rows.Next() {
		err := rows.Scan(&table)
		if err != nil {
			return version, tables, err
		}

		tables = append(tables, table)
	}

	if utils.IsStringInSlice(configTableName, tables) {
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

func (p *SQLProvider) upgrade() (err error) {
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
