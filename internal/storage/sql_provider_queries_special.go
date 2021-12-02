package storage

const (
	queryFmtDropTableIfExists = `DROP TABLE IF EXISTS %s;`

	queryFmtRenameTable = `
		ALTER TABLE %s
		RENAME TO %s;`

	queryFmtMySQLRenameTable = `
		ALTER TABLE %s
		RENAME %s;`
)

// Pre1 migration constants.
const (
	queryFmtPre1To1SelectAuthenticationLogs = `
		SELECT username, successful, time
		FROM %s
		ORDER BY time ASC
		LIMIT 100 OFFSET ?;`

	queryFmtPre1To1InsertAuthenticationLogs = `
		INSERT INTO %s (username, successful, time, request_uri)
		VALUES (?, ?, ?, '');`

	queryFmtPre1InsertUserPreferencesFromSelect = `
		INSERT INTO %s (username, second_factor_method)
		SELECT username, second_factor_method
		FROM %s
		ORDER BY username ASC;`

	queryFmtPre1SelectTOTPConfigurations = `
		SELECT username, secret
		FROM %s
		ORDER BY username ASC;`

	queryFmtPre1To1InsertTOTPConfiguration = `
		INSERT INTO %s (username, issuer, period, secret)
		VALUES (?, ?, ?, ?);`

	queryFmt1ToPre1InsertTOTPConfiguration = `
		INSERT INTO %s (username, secret)
		VALUES (?, ?);`

	queryFmtPre1To1SelectU2FDevices = `
		SELECT username, keyHandle, publicKey
		FROM %s
		ORDER BY username ASC;`

	queryFmtPre1To1InsertU2FDevice = `
		INSERT INTO %s (username, key_handle, public_key)
		VALUES (?, ?, ?);`

	queryFmt1ToPre1InsertAuthenticationLogs = `
		INSERT INTO %s (username, successful, time)
		VALUES (?, ?, ?);`

	queryFmt1ToPre1SelectAuthenticationLogs = `
		SELECT username, successful, time
		FROM %s
		ORDER BY id ASC
		LIMIT 100 OFFSET ?;`

	queryFmt1ToPre1SelectU2FDevices = `
		SELECT username, key_handle, public_key
		FROM %s
		ORDER BY username ASC;`

	queryFmt1ToPre1InsertU2FDevice = `
		INSERT INTO %s (username, keyHandle, publicKey)
		VALUES (?, ?, ?);`

	queryCreatePre1 = `
		CREATE TABLE user_preferences (
			username VARCHAR(100),
			second_factor_method VARCHAR(11),
			PRIMARY KEY (username)
		);
		
		CREATE TABLE identity_verification_tokens (
			token VARCHAR(512)
		);
		
		CREATE TABLE totp_secrets (
			username VARCHAR(100),
			secret VARCHAR(64),
			PRIMARY KEY (username)
		);
		
		CREATE TABLE u2f_devices (
			username VARCHAR(100),
			keyHandle TEXT,
			publicKey TEXT,
			PRIMARY KEY (username)
		);
		
		CREATE TABLE authentication_logs (
			username VARCHAR(100),
			successful BOOL,
			time INTEGER
		);
		
		CREATE TABLE config (
			category VARCHAR(32) NOT NULL,
			key_name VARCHAR(32) NOT NULL,
			value    TEXT,
			PRIMARY KEY (category, key_name)
		);
		
		INSERT INTO config (category, key_name, value)
		VALUES ('schema', 'version', '1');`
)
