package storage

const (
	queryFmtSelectLatestMigration = `
		SELECT id, applied, version_before, version_after, application_version
		FROM %s
		ORDER BY id DESC
		LIMIT 1;`

	queryFmtInsertMigration = `
		INSERT INTO %s (applied, version_before, version_after, application_version)
		VALUES (?, ?, ?, ?);`
)

const (
	queryMySQLSelectExistingTables = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_type = 'BASE TABLE' AND table_schema = database();`

	queryPostgreSelectExistingTables = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_type = 'BASE TABLE' AND table_schema = 'public';`

	querySQLiteSelectExistingTables = `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table';`
)

const (
	queryFmtSelectPreferred2FAMethod = `
		SELECT second_factor_method
		FROM %s
		WHERE username = ?;`

	queryFmtUpsertPreferred2FAMethod = `
		REPLACE INTO %s (username, second_factor_method)
		VALUES (?, ?);`

	queryFmtPostgresUpsertPreferred2FAMethod = `
		INSERT INTO %s (username, second_factor_method)
		VALUES ($1, $2)
			ON CONFLICT (username)
			DO UPDATE SET second_factor_method = $2;`
)

const (
	queryFmtSelectExistsIdentityVerification = `
		SELECT EXISTS (
			SELECT id
			FROM %s
			WHERE token = ?
		);`

	queryFmtInsertIdentityVerification = `
		INSERT INTO %s (token)
		VALUES (?);`

	queryFmtDeleteIdentityVerification = `
		DELETE FROM %s
		WHERE token = ?;`
)

const (
	queryFmtSelectTOTPConfiguration = `
		SELECT id, username, algorithm, digits, totp_period, secret
		FROM %s
		WHERE username = ?;`

	queryFmtUpsertTOTPConfiguration = `
		REPLACE INTO %s (username, algorithm, digits, totp_period, secret)
		VALUES (?, ?, ?, ?, ?);`

	queryFmtPostgresUpsertTOTPConfiguration = `
		INSERT INTO %s (username, algorithm, digits, totp_period, secret)
		VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (username)
			DO UPDATE SET algorithm = $2, digits = $3, totp_period = $4, secret = $5;`

	queryFmtDeleteTOTPConfiguration = `
		DELETE FROM %s
		WHERE username = ?;`
)

const (
	queryFmtSelectU2FDevice = `
		SELECT key_handle, public_key
		FROM %s
		WHERE username = ?;`

	queryFmtUpsertU2FDevice = `
		REPLACE INTO %s (username, key_handle, public_key)
		VALUES (?, ?, ?);`

	queryFmtPostgresUpsertU2FDevice = `
		INSERT INTO %s (username, key_handle, public_key)
		VALUES ($1, $2, $3)
			ON CONFLICT (username)
			DO UPDATE SET key_handle=$2, public_key=$3;`
)

const (
	queryFmtInsertAuthenticationLogEntry = `
		INSERT INTO %s (time, successful, username, remote_ip, request_uri, request_method)
		VALUES (?, ?, ?, ?, ?, ?);`

	queryFmtSelect1FAAuthenticationLogEntryByUsername = `
		SELECT time, successful, username
		FROM %s
		WHERE time > ? AND username = ? AND auth_type = '1FA'
		ORDER BY time DESC
		LIMIT ?
		OFFSET ?;`
)
