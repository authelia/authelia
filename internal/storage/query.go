package storage

const (
	queryFmtSelectLatestMigrationsVersion = `
		SELECT version
		FROM %s
		ORDER BY time DESC
		LIMIT 1;`

	queryRenameTable = `
		ALTER TABLE ?
		RENAME TO ?;`

	queryMySQLRenameTable = `
		ALTER TABLE ?
		RENAME ?;`
)

const (
	queryMySQLSelectExistingTables = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_type='BASE TABLE' AND table_schema=database();`

	queryPostgreSelectExistingTables = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_type='BASE TABLE' AND table_schema='public';`

	querySQLiteSelectExistingTables = `
		SELECT name
		FROM sqlite_master
		WHERE type='table';`
)

const (
	queryFmtSelectPreferred2FAMethodByUsername = `
		SELECT second_factor_method
		FROM %s
		WHERE username=?;`

	queryFmtUpsertPreferred2FAMethod = `
		REPLACE INTO %s (username, second_factor_method)
		VALUES (?, ?);`

	queryFmtPostgresUpsertPreferred2FAMethod = `
		INSERT INTO %s (username, second_factor_method)
		VALUES ($1, $2)
			ON CONFLICT (username)
			DO UPDATE SET second_factor_method=$2;`
)

const (
	// TODO: Select ID.
	queryFmtSelectExistsIdentityVerificationToken = `
		SELECT EXISTS (
			SELECT * FROM %s WHERE token=?
		);`

	queryFmtInsertIdentityVerificationToken = `
		INSERT INTO %s (token)
		VALUES (?);`

	queryFmtDeleteIdentityVerificationToken = `
		DELETE FROM %s
		WHERE token=?;`
)

const (
	queryFmtSelectTOTPSecretByUsername = `
		SELECT secret
		FROM %s
		WHERE username=?;`

	queryFmtUpsertTOTPSecret = `
		REPLACE INTO %s (username, secret)
		VALUES (?, ?);`

	queryFmtPostgresUpsertTOTPSecret = `
		INSERT INTO %s (username, secret)
		VALUES ($1, $2)
			ON CONFLICT (username)
			DO UPDATE SET secret=$2;`

	queryFmtDeleteTOTPSecret = `
		DELETE FROM %s
		WHERE username=?;`
)

const (
	queryFmtSelectU2FDeviceByUsername = `
		SELECT key_handle, public_key
		FROM %s
		WHERE username=?;`

	queryFmtUpsertU2FDevice = `
		REPLACE INTO %s (username, key_handle, public_key)
		VALUES (?, ?, ?);`

	queryFmtPostgresUpsertU2FDevice = `
		INSERT INTO %s (username, key_handle, public_key)
		VALUES ($1, $2, $3)
			ON CONFLICT (username)
			DO UPDATE SET key_handle=$2, public_key=$3`
)

const (
	queryFmtInsertAuthenticationAttempt = `
		INSERT INTO %s (username, successful, time)
		VALUES (?, ?, ?);`

	queryFmtSelectAuthenticationAttemptsByUsername = `
		SELECT username, successful, time
		FROM %s WHERE time>? AND username=?
		ORDER BY time DESC
		LIMIT ?
		OFFSET ?;`
)
