package storage

const (
	queryFmtSelectMigrations = `
		SELECT id, applied, version_before, version_after, application_version
		FROM %s;`

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
		WHERE table_type = 'BASE TABLE' AND table_schema = $1;`

	querySQLiteSelectExistingTables = `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table';`
)

const (
	queryFmtSelectUserInfo = `
		SELECT second_factor_method, (SELECT EXISTS (SELECT id FROM %s WHERE username = ?)) AS has_totp, (SELECT EXISTS (SELECT id FROM %s WHERE username = ?)) AS has_u2f, (SELECT EXISTS (SELECT id FROM %s WHERE username = ?)) AS has_duo
		FROM %s
		WHERE username = ?;`

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
	queryFmtSelectIdentityVerification = `
		SELECT id, jti, iat, issued_ip, exp, username, action, consumed, consumed_ip
		FROM %s
		WHERE jti = ?;`

	queryFmtInsertIdentityVerification = `
		INSERT INTO %s (jti, iat, issued_ip, exp, username, action)
		VALUES (?, ?, ?, ?, ?, ?);`

	queryFmtConsumeIdentityVerification = `
		UPDATE %s
		SET consumed = CURRENT_TIMESTAMP, consumed_ip = ?
		WHERE jti = ?;`
)

const (
	queryFmtSelectTOTPConfiguration = `
		SELECT id, username, issuer, algorithm, digits, period, secret
		FROM %s
		WHERE username = ?;`

	queryFmtSelectTOTPConfigurations = `
		SELECT id, username, issuer, algorithm, digits, period, secret
		FROM %s
		LIMIT ?
		OFFSET ?;`

	//nolint:gosec // These are not hardcoded credentials it's a query to obtain credentials.
	queryFmtUpdateTOTPConfigurationSecret = `
		UPDATE %s
		SET secret = ?
		WHERE id = ?;`

	//nolint:gosec // These are not hardcoded credentials it's a query to obtain credentials.
	queryFmtUpdateTOTPConfigurationSecretByUsername = `
		UPDATE %s
		SET secret = ?
		WHERE username = ?;`

	queryFmtUpsertTOTPConfiguration = `
		REPLACE INTO %s (username, issuer, algorithm, digits, period, secret)
		VALUES (?, ?, ?, ?, ?, ?);`

	queryFmtPostgresUpsertTOTPConfiguration = `
		INSERT INTO %s (username, issuer, algorithm, digits, period, secret)
		VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (username)
			DO UPDATE SET issuer = $2, algorithm = $3, digits = $4, period = $5, secret = $6;`

	queryFmtDeleteTOTPConfiguration = `
		DELETE FROM %s
		WHERE username = ?;`
)

const (
	queryFmtSelectU2FDevice = `
		SELECT id, username, key_handle, public_key
		FROM %s
		WHERE username = ?;`

	queryFmtSelectU2FDevices = `
		SELECT id, username, key_handle, public_key
		FROM %s
		LIMIT ?
		OFFSET ?;`

	queryFmtUpdateU2FDevicePublicKey = `
		UPDATE %s
		SET public_key = ?
		WHERE id = ?;`

	queryFmtUpdateUpdateU2FDevicePublicKeyByUsername = `
		UPDATE %s
		SET public_key = ?
		WHERE username = ?;`

	queryFmtUpsertU2FDevice = `
		REPLACE INTO %s (username, description, key_handle, public_key)
		VALUES (?, ?, ?, ?);`

	queryFmtPostgresUpsertU2FDevice = `
		INSERT INTO %s (username, description, key_handle, public_key)
		VALUES ($1, $2, $3, $4)
			ON CONFLICT (username, description)
			DO UPDATE SET key_handle=$3, public_key=$4;`
)

const (
	queryFmtUpsertDuoDevice = `
		REPLACE INTO %s (username, device, method)
		VALUES (?, ?, ?);`

	queryFmtPostgresUpsertDuoDevice = `
		INSERT INTO %s (username, device, method)
		VALUES ($1, $2, $3)
			ON CONFLICT (username)
			DO UPDATE SET device=$2, method=$3;`

	queryFmtDeleteDuoDevice = `
		DELETE
		FROM %s
		WHERE username = ?;`

	queryFmtSelectDuoDevice = `
		SELECT id, username, device, method
		FROM %s
		WHERE username = ?
		ORDER BY id;`
)

const (
	queryFmtInsertAuthenticationLogEntry = `
		INSERT INTO %s (time, successful, banned, username, auth_type, remote_ip, request_uri, request_method)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtSelect1FAAuthenticationLogEntryByUsername = `
		SELECT time, successful, username
		FROM %s
		WHERE time > ? AND username = ? AND auth_type = '1FA' AND banned = 0
		ORDER BY time DESC
		LIMIT ?
		OFFSET ?;`
)

const (
	queryFmtSelectEncryptionValue = `
		SELECT (value)
        FROM %s
        WHERE name = ?`

	queryFmtUpsertEncryptionValue = `
		REPLACE INTO %s (name, value)
		VALUES (?, ?);`

	queryFmtPostgresUpsertEncryptionValue = `
		INSERT INTO %s (name, value)
		VALUES ($1, $2)
			ON CONFLICT (name)
			DO UPDATE SET value=$2;`
)
