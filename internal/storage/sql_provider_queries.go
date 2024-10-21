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
		SELECT second_factor_method, (SELECT EXISTS (SELECT id FROM %s WHERE username = ?)) AS has_totp, (SELECT EXISTS (SELECT id FROM %s WHERE username = ?)) AS has_webauthn, (SELECT EXISTS (SELECT id FROM %s WHERE username = ?)) AS has_duo
		FROM %s
		WHERE username = ?;`

	queryFmtSelectPreferred2FAMethod = `
		SELECT second_factor_method
		FROM %s
		WHERE username = ?;`

	queryFmtUpsertPreferred2FAMethod = `
		REPLACE INTO %s (username, second_factor_method)
		VALUES (?, ?);`

	queryFmtUpsertPreferred2FAMethodPostgreSQL = `
		INSERT INTO %s (username, second_factor_method)
		VALUES ($1, $2)
			ON CONFLICT (username)
			DO UPDATE SET second_factor_method = $2;`
)

const (
	queryFmtSelectIdentityVerification = `
		SELECT id, jti, iat, issued_ip, exp, username, action, consumed, consumed_ip, revoked, revoked_ip
		FROM %s
		WHERE jti = ?;`

	queryFmtInsertIdentityVerification = `
		INSERT INTO %s (jti, iat, issued_ip, exp, username, action)
		VALUES (?, ?, ?, ?, ?, ?);`

	queryFmtConsumeIdentityVerification = `
		UPDATE %s
		SET consumed = CURRENT_TIMESTAMP, consumed_ip = ?
		WHERE jti = ?;`

	queryFmtRevokeIdentityVerification = `
		UPDATE %s
		SET revoked = CURRENT_TIMESTAMP, revoked_ip = ?
		WHERE jti = ?;`
)

const (
	queryFmtSelectOTCBySignatureAndUsername = `
		SELECT id, public_id, signature, issued, issued_ip, expires, username, intent, consumed, consumed_ip, revoked, revoked_ip, code
		FROM %s
		WHERE signature = ? AND username = ?;`

	queryFmtSelectOTCBySignature = `
		SELECT id, public_id, signature, issued, issued_ip, expires, username, intent, consumed, consumed_ip, revoked, revoked_ip, code
		FROM %s
		WHERE signature = ?;`

	queryFmtSelectOTCByID = `
		SELECT id, public_id, signature, issued, issued_ip, expires, username, intent, consumed, consumed_ip, revoked, revoked_ip, code
		FROM %s
		WHERE id = ?;`

	queryFmtSelectOTCByPublicID = `
		SELECT id, public_id, signature, issued, issued_ip, expires, username, intent, consumed, consumed_ip, revoked, revoked_ip, code
		FROM %s
		WHERE public_id = ?;`

	queryFmtInsertOTC = `
		INSERT INTO %s (public_id, signature, issued, issued_ip, expires, username, intent, code)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtConsumeOTC = `
		UPDATE %s
		SET consumed = ?, consumed_ip = ?
		WHERE signature = ?;`

	queryFmtRevokeOTC = `
		UPDATE %s
		SET revoked = CURRENT_TIMESTAMP, revoked_ip = ?
		WHERE public_id = ?;`

	queryFmtSelectOTCEncryptedData = `
		SELECT id, code
		FROM %s;`

	queryFmtUpdateOTCEncryptedData = `
		UPDATE %s
		SET code = ?
		WHERE id = ?;`
)

const (
	queryFmtSelectTOTPConfiguration = `
		SELECT id, created_at, last_used_at, username, issuer, algorithm, digits, period, secret
		FROM %s
		WHERE username = ?;`

	queryFmtSelectTOTPConfigurations = `
		SELECT id, created_at, last_used_at, username, issuer, algorithm, digits, period, secret
		FROM %s
		LIMIT ?
		OFFSET ?;`

	queryFmtUpsertTOTPConfiguration = `
		REPLACE INTO %s (created_at, last_used_at, username, issuer, algorithm, digits, period, secret)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtUpsertTOTPConfigurationPostgreSQL = `
		INSERT INTO %s (created_at, last_used_at, username, issuer, algorithm, digits, period, secret)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (username)
			DO UPDATE SET created_at = $1, last_used_at = $2, issuer = $4, algorithm = $5, digits = $6, period = $7, secret = $8;`

	queryFmtUpdateTOTPConfigRecordSignIn = `
		UPDATE %s
		SET last_used_at = ?
		WHERE id = ?;`

	queryFmtUpdateTOTPConfigRecordSignInByUsername = `
		UPDATE %s
		SET last_used_at = ?
		WHERE username = ?;`

	queryFmtDeleteTOTPConfiguration = `
		DELETE FROM %s
		WHERE username = ?;`

	queryFmtSelectTOTPConfigurationsEncryptedData = `
		SELECT id, secret
		FROM %s;`

	queryFmtUpdateTOTPConfigurationEncryptedData = `
		UPDATE %s
		SET secret = ?
		WHERE id = ?;`
)

const (
	queryFmtInsertTOTPHistory = `
		INSERT INTO %s (username, step)
		VALUES (?, ?);`

	queryFmtSelectTOTPHistory = `
		SELECT COUNT(id)
		FROM %s
		WHERE username = ? AND step = ?;`
)

//nolint:gosec // The following queries are not hard coded credentials.
const (
	queryFmtSelectWebAuthnCredentials = `
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key
		FROM %s
		LIMIT ?
		OFFSET ?;`

	queryFmtSelectWebAuthnCredentialsByUsername = `
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key
		FROM %s
		WHERE username = ?;`

	queryFmtSelectWebAuthnCredentialsByRPIDByUsername = `
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key
		FROM %s
		WHERE rpid = ? AND username = ?;`

	queryFmtSelectWebAuthnCredentialByID = `
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key
		FROM %s
		WHERE id = ?;`

	queryFmtUpdateUpdateWebAuthnCredentialDescriptionByUsernameAndID = `
		UPDATE %s
		SET description = ?
		WHERE username = ? AND id = ?;`

	queryFmtUpdateWebAuthnCredentialRecordSignIn = `
		UPDATE %s
		SET
			rpid = ?, last_used_at = ?, sign_count = ?, discoverable = ?, present = ?, verified = ?, backup_eligible = ?, backup_state = ?,
			clone_warning = CASE clone_warning WHEN TRUE THEN TRUE ELSE ? END
		WHERE id = ?;`

	queryFmtInsertWebAuthnCredential = `
		INSERT INTO %s (created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, discoverable, present, verified, backup_eligible, backup_state, public_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtDeleteWebAuthnCredential = `
		DELETE FROM %s
		WHERE kid = ?;`

	queryFmtDeleteWebAuthnCredentialByUsername = `
		DELETE FROM %s
		WHERE username = ?;`

	queryFmtDeleteWebAuthnCredentialByUsernameAndDescription = `
		DELETE FROM %s
		WHERE username = ? AND description = ?;`

	queryFmtSelectWebAuthnCredentialsEncryptedData = `
		SELECT id, public_key
		FROM %s;`

	queryFmtUpdateWebAuthnCredentialsEncryptedData = `
		UPDATE %s
		SET public_key = ?
		WHERE id = ?;`
)

const (
	queryFmtInsertWebAuthnUser = `
		INSERT INTO %s (rpid, username, userid)
		VALUES (?, ?, ?);`

	queryFmtSelectWebAuthnUser = `
		SELECT id, rpid, username, userid
		FROM %s
		WHERE rpid = ? AND username = ?;`
)

const (
	queryFmtUpsertDuoDevice = `
		REPLACE INTO %s (username, device, method)
		VALUES (?, ?, ?);`

	queryFmtUpsertDuoDevicePostgreSQL = `
		INSERT INTO %s (username, device, method)
		VALUES ($1, $2, $3)
			ON CONFLICT (username)
			DO UPDATE SET device = $2, method = $3;`

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
		WHERE time > ? AND username = ? AND auth_type = '1FA' AND banned = FALSE
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

	queryFmtUpsertEncryptionValuePostgreSQL = `
		INSERT INTO %s (name, value)
		VALUES ($1, $2)
			ON CONFLICT (name)
			DO UPDATE SET value = $2;`

	queryFmtSelectEncryptionEncryptedData = `
		SELECT id, value
		FROM %s;`

	queryFmtUpdateEncryptionEncryptedData = `
		UPDATE %s
		SET value = ?
		WHERE id = ?;`
)

const (
	queryFmtSelectOAuth2ConsentPreConfigurations = `
		SELECT id, client_id, subject, created_at, expires_at, revoked, scopes, audience
		FROM %s
		WHERE client_id = ? AND subject = ? AND
			  revoked = FALSE AND (expires_at IS NULL OR expires_at >= CURRENT_TIMESTAMP);`

	queryFmtInsertOAuth2ConsentPreConfiguration = `
		INSERT INTO %s (client_id, subject, created_at, expires_at, revoked, scopes, audience)
		VALUES(?, ?, ?, ?, ?, ?, ?);`

	queryFmtInsertOAuth2ConsentPreConfigurationPostgreSQL = `
		INSERT INTO %s (client_id, subject, created_at, expires_at, revoked, scopes, audience)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;`

	queryFmtSelectOAuth2ConsentSessionByChallengeID = `
		SELECT id, challenge_id, client_id, subject, authorized, granted, requested_at, responded_at,
		form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, preconfiguration
		FROM %s
		WHERE challenge_id = ?;`

	queryFmtInsertOAuth2ConsentSession = `
		INSERT INTO %s (challenge_id, client_id, subject, authorized, granted, requested_at, responded_at,
		form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, preconfiguration)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtUpdateOAuth2ConsentSessionSubject = `
		UPDATE %s
		SET subject = ?
		WHERE id = ?;`

	queryFmtUpdateOAuth2ConsentSessionResponse = `
		UPDATE %s
		SET authorized = ?, responded_at = CURRENT_TIMESTAMP, granted_scopes = ?, granted_audience = ?, preconfiguration = ?
		WHERE id = ? AND responded_at IS NULL;`

	queryFmtUpdateOAuth2ConsentSessionGranted = `
		UPDATE %s
		SET granted = TRUE
		WHERE id = ? AND responded_at IS NOT NULL;`

	queryFmtSelectOAuth2Session = `
		SELECT id, challenge_id, request_id, client_id, signature, subject, requested_at,
		requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data
		FROM %s
		WHERE signature = ? AND revoked = FALSE;`

	queryFmtInsertOAuth2Session = `
		INSERT INTO %s (challenge_id, request_id, client_id, signature, subject, requested_at,
		requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtRevokeOAuth2Session = `
		UPDATE %s
		SET revoked = TRUE
		WHERE signature = ?;`

	queryFmtRevokeOAuth2SessionByRequestID = `
		UPDATE %s
		SET revoked = TRUE
		WHERE request_id = ?;`

	queryFmtDeactivateOAuth2Session = `
		UPDATE %s
		SET active = FALSE
		WHERE signature = ?;`

	queryFmtDeactivateOAuth2SessionByRequestID = `
		UPDATE %s
		SET active = FALSE
		WHERE request_id = ?;`

	queryFmtSelectOAuth2DeviceCodeSession = `
		SELECT id, challenge_id, request_id, client_id, signature, user_code_signature, status, subject,
		requested_at, checked_at, requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data
		FROM %s
		WHERE signature = ? AND revoked = FALSE;`

	queryFmtSelectOAuth2DeviceCodeSessionByUserCode = `
		SELECT id, challenge_id, request_id, client_id, signature, user_code_signature, status, subject,
		requested_at, checked_at, requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data
		FROM %s
		WHERE user_code_signature = ? AND revoked = FALSE;`

	queryFmtInsertOAuth2DeviceCodeSession = `
		INSERT INTO %s (challenge_id, request_id, client_id, signature, user_code_signature, status, subject,
		requested_at, checked_at, requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtUpdateOAuth2DeviceCodeSession = `
		UPDATE %s
		SET checked_at = ?, status = ?
		WHERE signature = ?;`

	queryFmtSelectOAuth2PARContext = `
		SELECT id, signature, request_id, client_id, requested_at, scopes, audience,
		handled_response_types, response_mode, response_mode_default, revoked,
		form_data, session_data
		FROM %s
		WHERE signature = ?;`

	queryFmtInsertOAuth2PARContext = `
		INSERT INTO %s (signature, request_id, client_id, requested_at, scopes, audience,
		handled_response_types, response_mode, response_mode_default, revoked,
		form_data, session_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtUpdateOAuth2PARContext = `
	UPDATE %s
	SET signature = ?, request_id = ?, client_id = ?, requested_at = ?, scopes = ?, audience = ?,
	    handled_response_types = ?, response_mode = ?, response_mode_default = ?, revoked = ?,
	    form_data = ?, session_data = ?
	WHERE id = ?;`

	queryFmtSelectOAuth2BlacklistedJTI = `
		SELECT id, signature, expires_at
		FROM %s
		WHERE signature = ?;`

	queryFmtUpsertOAuth2BlacklistedJTI = `
		REPLACE INTO %s (signature, expires_at)
		VALUES(?, ?);`

	queryFmtUpsertOAuth2BlacklistedJTIPostgreSQL = `
		INSERT INTO %s (signature, expires_at)
		VALUES ($1, $2)
			ON CONFLICT (signature)
			DO UPDATE SET expires_at = $2;`

	queryFmtSelectOAuth2SessionEncryptedData = `
		SELECT id, session_data
		FROM %s;`

	queryFmtUpdateOAuth2ConsentSessionEncryptedData = `
		UPDATE %s
		SET session_data = ?
		WHERE id = ?;`
)

const (
	queryFmtInsertUserOpaqueIdentifier = `
		INSERT INTO %s (service, sector_id, username, identifier)
		VALUES(?, ?, ?, ?);`

	queryFmtSelectUserOpaqueIdentifier = `
		SELECT id, service, sector_id, username, identifier
		FROM %s
		WHERE identifier = ?;`

	queryFmtSelectUserOpaqueIdentifierBySignature = `
		SELECT id, service, sector_id, username, identifier
		FROM %s
		WHERE service = ? AND sector_id = ? AND username = ?;`

	queryFmtSelectUserOpaqueIdentifiers = `
		SELECT id, service, sector_id, username, identifier
		FROM %s;`
)
