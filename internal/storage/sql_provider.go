package storage

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/authelia/authelia/internal/models"
)

// SQLProvider is a storage provider persisting data in a SQL database.
type SQLProvider struct {
	db *sql.DB

	sqlCreateUserPreferencesTable            string
	sqlCreateIdentityVerificationTokensTable string
	sqlCreateTOTPSecretsTable                string
	sqlCreateU2FDeviceHandlesTable           string
	sqlCreateAuthenticationLogsTable         string
	sqlCreateAuthenticationLogsUserTimeIndex string

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

	sqlInsertAuthenticationLog     string
	sqlGetLatestAuthenticationLogs string
}

func (p *SQLProvider) initialize(db *sql.DB) error {
	p.db = db

	_, err := db.Exec(p.sqlCreateUserPreferencesTable)
	if err != nil {
		return fmt.Errorf("Unable to create table %s: %v", preferencesTableName, err)
	}

	_, err = db.Exec(p.sqlCreateIdentityVerificationTokensTable)
	if err != nil {
		return fmt.Errorf("Unable to create table %s: %v", identityVerificationTokensTableName, err)
	}

	_, err = db.Exec(p.sqlCreateTOTPSecretsTable)
	if err != nil {
		return fmt.Errorf("Unable to create table %s: %v", totpSecretsTableName, err)
	}

	// keyHandle and publicKey are stored in base64 format
	_, err = db.Exec(p.sqlCreateU2FDeviceHandlesTable)
	if err != nil {
		return fmt.Errorf("Unable to create table %s: %v", u2fDeviceHandlesTableName, err)
	}

	_, err = db.Exec(p.sqlCreateAuthenticationLogsTable)
	if err != nil {
		return fmt.Errorf("Unable to create table %s: %v", authenticationLogsTableName, err)
	}

	// Create an index on (username, time) because this couple is highly used by the regulation module
	// to check whether a user is banned.
	if p.sqlCreateAuthenticationLogsUserTimeIndex != "" {
		_, err = db.Exec(p.sqlCreateAuthenticationLogsUserTimeIndex)
		if err != nil {
			return fmt.Errorf("Unable to create table %s: %v", authenticationLogsTableName, err)
		}
	}
	return nil
}

// LoadPreferred2FAMethod load the preferred method for 2FA from sqlite db.
func (p *SQLProvider) LoadPreferred2FAMethod(username string) (string, error) {
	rows, err := p.db.Query(p.sqlGetPreferencesByUsername, username)
	defer rows.Close()
	if err != nil {
		return "", err
	}
	if rows.Next() {
		var method string
		err = rows.Scan(&method)
		if err != nil {
			return "", err
		}
		return method, nil
	}
	return "", nil
}

// SavePreferred2FAMethod save the preferred method for 2FA in sqlite db.
func (p *SQLProvider) SavePreferred2FAMethod(username string, method string) error {
	_, err := p.db.Exec(p.sqlUpsertSecondFactorPreference, username, method)
	return err
}

// FindIdentityVerificationToken look for an identity verification token in DB.
func (p *SQLProvider) FindIdentityVerificationToken(token string) (bool, error) {
	var found bool
	err := p.db.QueryRow(p.sqlTestIdentityVerificationTokenExistence, token).Scan(&found)
	if err != nil {
		return false, err
	}
	return found, nil
}

// SaveIdentityVerificationToken save an identity verification token in DB.
func (p *SQLProvider) SaveIdentityVerificationToken(token string) error {
	_, err := p.db.Exec(p.sqlInsertIdentityVerificationToken, token)
	return err
}

// RemoveIdentityVerificationToken remove an identity verification token from the DB.
func (p *SQLProvider) RemoveIdentityVerificationToken(token string) error {
	_, err := p.db.Exec(p.sqlDeleteIdentityVerificationToken, token)
	return err
}

// SaveTOTPSecret save a TOTP secret of a given user.
func (p *SQLProvider) SaveTOTPSecret(username string, secret string) error {
	_, err := p.db.Exec(p.sqlUpsertTOTPSecret, username, secret)
	return err
}

// LoadTOTPSecret load a TOTP secret given a username.
func (p *SQLProvider) LoadTOTPSecret(username string) (string, error) {
	var secret string
	if err := p.db.QueryRow(p.sqlGetTOTPSecretByUsername, username).Scan(&secret); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoTOTPSecret
		}
		return "", err
	}
	return secret, nil
}

// DeleteTOTPSecret delete a TOTP secret given a username.
func (p *SQLProvider) DeleteTOTPSecret(username string) error {
	_, err := p.db.Exec(p.sqlDeleteTOTPSecret, username)
	return err
}

// SaveU2FDeviceHandle save a registered U2F device registration blob.
func (p *SQLProvider) SaveU2FDeviceHandle(username string, keyHandle []byte, publicKey []byte) error {
	_, err := p.db.Exec(p.sqlUpsertU2FDeviceHandle,
		username,
		base64.StdEncoding.EncodeToString(keyHandle),
		base64.StdEncoding.EncodeToString(publicKey))
	return err
}

// LoadU2FDeviceHandle load a U2F device registration blob for a given username.
func (p *SQLProvider) LoadU2FDeviceHandle(username string) ([]byte, []byte, error) {
	var keyHandleBase64, publicKeyBase64 string
	if err := p.db.QueryRow(p.sqlGetU2FDeviceHandleByUsername, username).Scan(&keyHandleBase64, &publicKeyBase64); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrNoU2FDeviceHandle
		}
		return nil, nil, err
	}

	keyHandle, err := base64.StdEncoding.DecodeString(keyHandleBase64)

	if err != nil {
		return nil, nil, err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)

	if err != nil {
		return nil, nil, err
	}

	return keyHandle, publicKey, nil
}

// AppendAuthenticationLog append a mark to the authentication log.
func (p *SQLProvider) AppendAuthenticationLog(attempt models.AuthenticationAttempt) error {
	_, err := p.db.Exec(p.sqlInsertAuthenticationLog, attempt.Username, attempt.Successful, attempt.Time.Unix())
	return err
}

// LoadLatestAuthenticationLogs retrieve the latest marks from the authentication log.
func (p *SQLProvider) LoadLatestAuthenticationLogs(username string, fromDate time.Time) ([]models.AuthenticationAttempt, error) {
	rows, err := p.db.Query(p.sqlGetLatestAuthenticationLogs, fromDate.Unix(), username)

	if err != nil {
		return nil, err
	}

	attempts := make([]models.AuthenticationAttempt, 0, 10)
	for rows.Next() {
		attempt := models.AuthenticationAttempt{
			Username: username,
		}
		var t int64
		err = rows.Scan(&attempt.Successful, &t)
		attempt.Time = time.Unix(t, 0)

		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}
	return attempts, nil
}
