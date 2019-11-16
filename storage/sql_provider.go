package storage

import (
	"database/sql"
	"time"

	"github.com/clems4ever/authelia/models"
)

// SQLProvider is a storage provider persisting data in a SQL database.
type SQLProvider struct {
	db *sql.DB
}

func (p *SQLProvider) initialize(db *sql.DB) error {
	p.db = db

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS SecondFactorPreferences (username VARCHAR(100) PRIMARY KEY, method VARCHAR(10))")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS IdentityVerificationTokens (token VARCHAR(512))")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS TOTPSecrets (username VARCHAR(100) PRIMARY KEY, secret VARCHAR(64))")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS U2FDeviceHandles (username VARCHAR(100) PRIMARY KEY, deviceHandle BLOB)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS AuthenticationLogs (username VARCHAR(100), successful BOOL, time INTEGER)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS time ON AuthenticationLogs (time);")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS username ON AuthenticationLogs (username);")
	if err != nil {
		return err
	}
	return nil
}

// LoadPrefered2FAMethod load the prefered method for 2FA from sqlite db.
func (p *SQLProvider) LoadPrefered2FAMethod(username string) (string, error) {
	stmt, err := p.db.Prepare("SELECT method FROM SecondFactorPreferences WHERE username=?")
	if err != nil {
		return "", err
	}
	rows, err := stmt.Query(username)
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

// SavePrefered2FAMethod save the prefered method for 2FA in sqlite db.
func (p *SQLProvider) SavePrefered2FAMethod(username string, method string) error {
	stmt, err := p.db.Prepare("REPLACE INTO SecondFactorPreferences (username, method) VALUES (?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(username, method)
	return err
}

// FindIdentityVerificationToken look for an identity verification token in DB.
func (p *SQLProvider) FindIdentityVerificationToken(token string) (bool, error) {
	stmt, err := p.db.Prepare("SELECT token FROM IdentityVerificationTokens WHERE token=?")
	if err != nil {
		return false, err
	}
	var found string
	err = stmt.QueryRow(token).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SaveIdentityVerificationToken save an identity verification token in DB.
func (p *SQLProvider) SaveIdentityVerificationToken(token string) error {
	stmt, err := p.db.Prepare("INSERT INTO IdentityVerificationTokens (token) VALUES (?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(token)
	return err
}

// RemoveIdentityVerificationToken remove an identity verification token from the DB.
func (p *SQLProvider) RemoveIdentityVerificationToken(token string) error {
	stmt, err := p.db.Prepare("DELETE FROM IdentityVerificationTokens WHERE token=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(token)
	return err
}

// SaveTOTPSecret save a TOTP secret of a given user.
func (p *SQLProvider) SaveTOTPSecret(username string, secret string) error {
	stmt, err := p.db.Prepare("REPLACE INTO TOTPSecrets (username, secret) VALUES (?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(username, secret)
	return err
}

// LoadTOTPSecret load a TOTP secret given a username.
func (p *SQLProvider) LoadTOTPSecret(username string) (string, error) {
	stmt, err := p.db.Prepare("SELECT secret FROM TOTPSecrets WHERE username=?")
	if err != nil {
		return "", err
	}
	var secret string
	err = stmt.QueryRow(username).Scan(&secret)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return secret, nil
}

// SaveU2FDeviceHandle save a registered U2F device registration blob.
func (p *SQLProvider) SaveU2FDeviceHandle(username string, keyHandle []byte) error {
	stmt, err := p.db.Prepare("REPLACE INTO U2FDeviceHandles (username, deviceHandle) VALUES (?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(username, keyHandle)
	return err
}

// LoadU2FDeviceHandle load a U2F device registration blob for a given username.
func (p *SQLProvider) LoadU2FDeviceHandle(username string) ([]byte, error) {
	stmt, err := p.db.Prepare("SELECT deviceHandle FROM U2FDeviceHandles WHERE username=?")
	if err != nil {
		return nil, err
	}
	var deviceHandle []byte
	err = stmt.QueryRow(username).Scan(&deviceHandle)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoU2FDeviceHandle
		}
		return nil, err
	}
	return deviceHandle, nil
}

// AppendAuthenticationLog append a mark to the authentication log.
func (p *SQLProvider) AppendAuthenticationLog(attempt models.AuthenticationAttempt) error {
	stmt, err := p.db.Prepare("INSERT INTO AuthenticationLogs (username, successful, time) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(attempt.Username, attempt.Successful, attempt.Time.Unix())
	return err
}

// LoadLatestAuthenticationLogs retrieve the latest marks from the authentication log.
func (p *SQLProvider) LoadLatestAuthenticationLogs(username string, fromDate time.Time) ([]models.AuthenticationAttempt, error) {
	rows, err := p.db.Query("SELECT successful, time FROM AuthenticationLogs WHERE time>? AND username=? ORDER BY time DESC",
		fromDate.Unix(), username)

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
