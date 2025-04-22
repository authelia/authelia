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

	queryMSSQLSelectExistingTables = `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_SCHEMA = SCHEMA_NAME();`

	queryPostgreSelectExistingTables = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_type = 'BASE TABLE' AND table_schema = '%s';`

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

	queryFmtUpsertPreferred2FAMethodMSSQL = `
		BEGIN TRY
			INSERT INTO %s ([username], [second_factor_method])
			VALUES (@p1, @p2);
		END TRY
		BEGIN CATCH
			IF ERROR_NUMBER() IN (2601, 2627)
				UPDATE %s
				SET
					[second_factor_method] = @p2
				WHERE [username] = @p1;
		END CATCH;`
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

	queryFmtUpsertTOTPConfigurationMSSQL = `
		BEGIN TRY
			INSERT INTO %s ([created_at], [last_used_at], [username], [issuer], [algorithm], [digits], [period], [secret])
			VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8);
		END TRY
		BEGIN CATCH
			IF ERROR_NUMBER() IN (2601, 2627)
				UPDATE %s
				SET
					[created_at] = @p1,
					[last_used_at] = @p2,
					[issuer] = @p4,
					[algorithm] = @p5,
					[digits] = @p6,
					[period] = @p7,
					[secret] = @p8
				WHERE [username] = @p3;
		END CATCH;`

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
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key, attestation
		FROM %s
		LIMIT ?
		OFFSET ?;`

	queryFmtSelectWebAuthnCredentialsByUsername = `
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key, attestation
		FROM %s
		WHERE username = ? AND (? = 0 OR discoverable = 1);`

	queryFmtSelectWebAuthnCredentialsByRPIDByUsername = `
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key, attestation
		FROM %s
		WHERE rpid = ? AND username = ? AND (? = 0 OR discoverable = 1);`

	queryFmtSelectWebAuthnCredentialByID = `
		SELECT id, created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key, attestation
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
			clone_warning = CASE clone_warning WHEN 1 THEN 1 ELSE ? END
		WHERE id = ?;`

	queryFmtInsertWebAuthnCredential = `
		INSERT INTO %s (created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, discoverable, present, verified, backup_eligible, backup_state, public_key, attestation)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

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
		SELECT id, public_key, attestation
		FROM %s;`

	queryFmtUpdateWebAuthnCredentialsEncryptedData = `
		UPDATE %s
		SET public_key = ?, attestation = ?
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

	queryFmtSelectWebAuthnUserByUserID = `
		SELECT id, rpid, username, userid
		FROM %s
		WHERE rpid = ? AND userid = ?;`
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

	queryFmtUpsertDuoDeviceMSSQL = `
		BEGIN TRY
			INSERT INTO %s ([username], [device], [method])
			VALUES (@p1, @p2, @p3);
		END TRY
		BEGIN CATCH
			IF ERROR_NUMBER() IN (2601, 2627)
				UPDATE %s
				SET
					[device] = @p2,
					[method] = @p3
				WHERE [username] = @p1;
		END CATCH;`

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

	queryFmtSelectAuthenticationLogsRegulationRecordsByUsername = `
		SELECT time, successful
		FROM %s
		WHERE time > ? AND username = ? AND auth_type = '1FA' AND banned = ?
		ORDER BY time DESC
		LIMIT ?;`

	queryFmtSelectAuthenticationLogsRegulationRecordsByRemoteIP = `
		SELECT time, successful
		FROM %s
		WHERE time > ? AND remote_ip = ? AND auth_type = '1FA' AND banned = ?
		ORDER BY time DESC
		LIMIT ?;`
)

const (
	queryFmtInsertBannedUser = `
		INSERT INTO %s (expires, username, source, reason)
		VALUES (?, ?, ?, ?);`

	queryFmtSelectBannedUser = `
		SELECT id, time, expires, expired, revoked, username, source, reason
		FROM %s
		WHERE username = ? AND revoked = 0 AND (expires IS NULL OR expires > ?) AND expired IS NULL
		ORDER BY time DESC;`

	queryFmtSelectBannedUserByID = `
		SELECT id, time, expires, expired, revoked, username, source, reason
		FROM %s
		WHERE id = ?;`

	queryFmtSelectBannedUsers = `
		SELECT id, time, expires, expired, revoked, username, source, reason
		FROM %s
		WHERE revoked = ? AND (expires IS NULL OR expires > ?) AND expired IS NULL
		LIMIT ?
		OFFSET ?;`

	queryFmtSelectBannedUserLastExpires = `
		SELECT expires, expired, revoked
		FROM %s
		WHERE username = ?
		ORDER BY time DESC
		LIMIT 1;`
)

const (
	queryFmtInsertBannedIP = `
		INSERT INTO %s (expires, ip, source, reason)
		VALUES (?, ?, ?, ?);`

	queryFmtSelectBannedIP = `
		SELECT id, time, expires, expired, revoked, ip, source, reason
		FROM %s
		WHERE ip = ? AND revoked = ? AND (expires IS NULL OR expires > ?) AND expired IS NULL
		ORDER BY time DESC;`

	queryFmtSelectBannedIPByID = `
		SELECT id, time, expires, expired, revoked, ip, source, reason
		FROM %s
		WHERE id = ?;`

	queryFmtSelectBannedIPs = `
		SELECT id, time, expires, expired, revoked, ip, source, reason
		FROM %s
		WHERE revoked = 0 AND (expires IS NULL OR expires > ?)
		LIMIT ?
		OFFSET ?;`

	queryFmtSelectBannedIPLastExpires = `
		SELECT expires, expired, revoked
		FROM %s
		WHERE ip = ?
		ORDER BY time DESC
		LIMIT 1;`
)

const (
	queryFmtRevokeBannedEntry = `
		UPDATE %s
		SET expired = ?, revoked = 1
		WHERE id = ?;`
)

const (
	queryFmtUpsertCachedData = `
		REPLACE INTO %s (updated_at, name, encrypted, value)
		VALUES (CURRENT_TIMESTAMP, ?, ?, ?);`

	queryFmtUpsertCachedDataPostgreSQL = `
		INSERT INTO %s (updated_at, name, encrypted, value)
		VALUES (CURRENT_TIMESTAMP, $1, $2, $3)
			ON CONFLICT (name)
			DO UPDATE SET updated_at = CURRENT_TIMESTAMP, encrypted = $2, value = $3;`

	queryFmtUpsertCachedDataMSSQL = `
		BEGIN TRY
			INSERT INTO %s ([updated_at], [name], [encrypted], [value])
			VALUES (CURRENT_TIMESTAMP, @p1, @p2, @p3);
		END TRY
		BEGIN CATCH
			IF ERROR_NUMBER() IN (2601, 2627)
				UPDATE %s
				SET
					[updated_at] = CURRENT_TIMESTAMP,
					[encrypted] = @p2,
					[value] = @p3
				WHERE [name] = @p1;
		END CATCH;`

	queryFmtSelectCachedData = `
		SELECT id, created_at, updated_at, name, encrypted, value
		FROM %s
		WHERE name = ?;`

	queryFmtDeleteCachedData = `
		DELETE FROM %s
		WHERE name = ?;`

	queryFmtSelectCachedDataEncryptedData = `
		SELECT id, value
		FROM %s
		WHERE encrypted = 1;`

	queryFmtUpdateCachedDataEncryptedData = `
		UPDATE %s
		SET value = ?
		WHERE id = ?;`
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

	queryFmtUpsertEncryptionValueMSSQL = `
		BEGIN TRY
			INSERT INTO %s ([name], [value])
			VALUES (@p1, @p2);
		END TRY
		BEGIN CATCH
			IF ERROR_NUMBER() IN (2601, 2627)
				UPDATE %s
				SET
					[value] = @p2
				WHERE [name] = @p1;
		END CATCH;`

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
		SELECT id, client_id, subject, created_at, expires_at, revoked, scopes, audience, requested_claims, signature_claims, granted_claims
		FROM %s
		WHERE client_id = ? AND subject = ? AND
			  revoked = 0 AND (expires_at IS NULL OR expires_at >= CURRENT_TIMESTAMP);`

	queryFmtInsertOAuth2ConsentPreConfiguration = `
		INSERT INTO %s (client_id, subject, created_at, expires_at, revoked, scopes, audience, requested_claims, signature_claims, granted_claims)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtInsertOAuth2ConsentPreConfigurationPostgreSQL = `
		INSERT INTO %s (client_id, subject, created_at, expires_at, revoked, scopes, audience, requested_claims, signature_claims, granted_claims)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id;`

	queryFmtSelectOAuth2ConsentSessionByChallengeID = `
		SELECT id, challenge_id, client_id, subject, authorized, granted, requested_at, responded_at,
		form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, granted_claims, preconfiguration
		FROM %s
		WHERE challenge_id = ?;`

	queryFmtInsertOAuth2ConsentSession = `
		INSERT INTO %s (challenge_id, client_id, subject, authorized, granted, requested_at, responded_at,
		form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, granted_claims, preconfiguration)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtUpdateOAuth2ConsentSessionSubject = `
		UPDATE %s
		SET subject = ?
		WHERE id = ?;`

	queryFmtUpdateOAuth2ConsentSessionResponse = `
		UPDATE %s
		SET authorized = ?, responded_at = CURRENT_TIMESTAMP, granted_scopes = ?, granted_audience = ?, granted_claims = ?, preconfiguration = ?
		WHERE id = ? AND responded_at IS NULL;`

	queryFmtUpdateOAuth2ConsentSessionGranted = `
		UPDATE %s
		SET granted = 1
		WHERE id = ? AND responded_at IS NOT NULL;`

	queryFmtSelectOAuth2Session = `
		SELECT id, challenge_id, request_id, client_id, signature, subject, requested_at,
		requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data
		FROM %s
		WHERE signature = ? AND revoked = 0;`

	queryFmtInsertOAuth2Session = `
		INSERT INTO %s (challenge_id, request_id, client_id, signature, subject, requested_at,
		requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	queryFmtRevokeOAuth2Session = `
		UPDATE %s
		SET revoked = 1
		WHERE signature = ?;`

	queryFmtRevokeOAuth2SessionByRequestID = `
		UPDATE %s
		SET revoked = 1
		WHERE request_id = ?;`

	queryFmtDeactivateOAuth2Session = `
		UPDATE %s
		SET active = 0
		WHERE signature = ?;`

	queryFmtDeactivateOAuth2SessionByRequestID = `
		UPDATE %s
		SET active = 0
		WHERE request_id = ?;`

	queryFmtSelectOAuth2DeviceCodeSession = `
		SELECT id, challenge_id, request_id, client_id, signature, user_code_signature, status, subject,
		requested_at, checked_at, requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data
		FROM %s
		WHERE signature = ? AND revoked = 0;`

	queryFmtSelectOAuth2DeviceCodeSessionByUserCode = `
		SELECT id, challenge_id, request_id, client_id, signature, user_code_signature, status, subject,
		requested_at, checked_at, requested_scopes, granted_scopes, requested_audience, granted_audience,
		active, revoked, form_data, session_data
		FROM %s
		WHERE user_code_signature = ? AND revoked = 0;`

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

	queryFmtUpsertOAuth2BlacklistedJTIMSSQL = `
		BEGIN TRY
			INSERT INTO %s ([signature], [expires_at])
			VALUES (@p1, @p2);
		END TRY
		BEGIN CATCH
			IF ERROR_NUMBER() IN (2601, 2627)
				UPDATE %s
				SET
					[expires_at] = @p2
				WHERE [signature] = @p1;
		END CATCH;`

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
