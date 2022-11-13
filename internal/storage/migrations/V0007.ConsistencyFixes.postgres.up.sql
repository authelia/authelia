DROP TABLE IF EXISTS _bkp_UP_V0002_totp_configurations;
DROP TABLE IF EXISTS _bkp_UP_V0002_u2f_devices;
DROP TABLE IF EXISTS totp_secrets;
DROP TABLE IF EXISTS identity_verification_tokens;
DROP TABLE IF EXISTS u2f_devices;
DROP TABLE IF EXISTS config;
DROP TABLE IF EXISTS AuthenticationLogs;
DROP TABLE IF EXISTS IdentityVerificationTokens;
DROP TABLE IF EXISTS Preferences;
DROP TABLE IF EXISTS PreferencesTableName;
DROP TABLE IF EXISTS SecondFactorPreferences;
DROP TABLE IF EXISTS TOTPSecrets;
DROP TABLE IF EXISTS U2FDeviceHandles;

ALTER TABLE webauthn_devices
	ALTER COLUMN aaguid DROP NOT NULL;

UPDATE webauthn_devices
SET aaguid = NULL
WHERE aaguid = '' OR aaguid = '00000000-00000000-00000000-00000000';

ALTER TABLE duo_devices
	DROP CONSTRAINT IF EXISTS duo_devices_username_key;

DROP INDEX IF EXISTS duo_devices_username_key;

CREATE UNIQUE INDEX duo_devices_username_key ON duo_devices (username);

ALTER TABLE encryption
	DROP CONSTRAINT IF EXISTS encryption_name_key;

DROP INDEX IF EXISTS encryption_name_key;

CREATE UNIQUE INDEX encryption_name_key ON encryption (name);

ALTER TABLE identity_verification
	DROP CONSTRAINT IF EXISTS identity_verification_jti_key;

DROP INDEX IF EXISTS identity_verification_jti_key;

CREATE UNIQUE INDEX identity_verification_jti_key ON identity_verification (jti);

ALTER TABLE user_preferences
	DROP CONSTRAINT IF EXISTS user_preferences_username_key;

DROP INDEX IF EXISTS user_preferences_username_key;

CREATE UNIQUE INDEX user_preferences_username_key ON user_preferences (username);

ALTER TABLE totp_configurations
	DROP CONSTRAINT IF EXISTS totp_configurations_username_key1,
	DROP CONSTRAINT IF EXISTS totp_configurations_username_key,
	DROP CONSTRAINT IF EXISTS totp_configurations_pkey,
	DROP CONSTRAINT IF EXISTS totp_configurations_pkey1;

DROP INDEX IF EXISTS totp_configurations_username_key1;
DROP INDEX IF EXISTS totp_configurations_username_key;

ALTER TABLE totp_configurations
	RENAME TO _bkp_UP_V0007_totp_configurations;

CREATE TABLE IF NOT EXISTS totp_configurations (
	id SERIAL CONSTRAINT totp_configurations_pkey PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_used_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
	username VARCHAR(100) NOT NULL,
	issuer VARCHAR(100),
	algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
	digits INTEGER NOT NULL DEFAULT 6,
	period INTEGER NOT NULL DEFAULT 30,
	secret BYTEA NOT NULL
);

CREATE UNIQUE INDEX totp_configurations_username_key ON totp_configurations (username);

INSERT INTO totp_configurations (created_at, last_used_at, username, issuer, algorithm, digits, period, secret)
SELECT created_at, last_used_at, username, issuer, algorithm, digits, period, secret
FROM _bkp_UP_V0007_totp_configurations
ORDER BY id;

DROP TABLE IF EXISTS _bkp_UP_V0007_totp_configurations;

ALTER TABLE webauthn_devices
	DROP CONSTRAINT IF EXISTS webauthn_devices_username_description_key1,
	DROP CONSTRAINT IF EXISTS webauthn_devices_kid_key1,
	DROP CONSTRAINT IF EXISTS webauthn_devices_lookup_key1,
	DROP CONSTRAINT IF EXISTS webauthn_devices_username_description_key,
	DROP CONSTRAINT IF EXISTS webauthn_devices_kid_key,
	DROP CONSTRAINT IF EXISTS webauthn_devices_lookup_key,
	DROP CONSTRAINT IF EXISTS webauthn_devices_pkey,
	DROP CONSTRAINT IF EXISTS webauthn_devices_pkey1;

DROP INDEX IF EXISTS webauthn_devices_username_description_key1;
DROP INDEX IF EXISTS webauthn_devices_kid_key1;
DROP INDEX IF EXISTS webauthn_devices_lookup_key1;
DROP INDEX IF EXISTS webauthn_devices_username_description_key;
DROP INDEX IF EXISTS webauthn_devices_kid_key;
DROP INDEX IF EXISTS webauthn_devices_lookup_key;

ALTER TABLE webauthn_devices
	RENAME TO _bkp_UP_V0007_webauthn_devices;

CREATE TABLE IF NOT EXISTS webauthn_devices (
	id SERIAL CONSTRAINT webauthn_devices_pkey PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_used_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
	rpid TEXT,
	username VARCHAR(100) NOT NULL,
	description VARCHAR(30) NOT NULL DEFAULT 'Primary',
	kid VARCHAR(512) NOT NULL,
	public_key BYTEA NOT NULL,
	attestation_type VARCHAR(32),
	transport VARCHAR(20) DEFAULT '',
	aaguid CHAR(36) NOT NULL,
	sign_count INTEGER DEFAULT 0,
	clone_warning BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX webauthn_devices_kid_key ON webauthn_devices (kid);
CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (username, description);

INSERT INTO webauthn_devices (created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning)
SELECT created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning
FROM _bkp_UP_V0007_webauthn_devices;

DROP TABLE IF EXISTS _bkp_UP_V0007_webauthn_devices;

ALTER TABLE oauth2_consent_session
	DROP CONSTRAINT oauth2_consent_session_subject_fkey,
	DROP CONSTRAINT oauth2_consent_session_preconfiguration_fkey;

ALTER TABLE oauth2_consent_preconfiguration
	DROP CONSTRAINT IF EXISTS oauth2_consent_preconfiguration_subjct_fkey,
	DROP CONSTRAINT IF EXISTS  oauth2_consent_preconfiguration_subject_fkey;

ALTER TABLE oauth2_access_token_session
	DROP CONSTRAINT oauth2_access_token_session_challenge_id_fkey,
	DROP CONSTRAINT oauth2_access_token_session_subject_fkey;

ALTER TABLE oauth2_authorization_code_session
	DROP CONSTRAINT oauth2_authorization_code_session_challenge_id_fkey,
	DROP CONSTRAINT oauth2_authorization_code_session_subject_fkey;

ALTER TABLE oauth2_openid_connect_session
	DROP CONSTRAINT oauth2_openid_connect_session_challenge_id_fkey,
	DROP CONSTRAINT oauth2_openid_connect_session_subject_fkey;

ALTER TABLE oauth2_pkce_request_session
	DROP CONSTRAINT oauth2_pkce_request_session_challenge_id_fkey,
	DROP CONSTRAINT oauth2_pkce_request_session_subject_fkey;

ALTER TABLE oauth2_refresh_token_session
	DROP CONSTRAINT oauth2_refresh_token_session_challenge_id_fkey,
	DROP CONSTRAINT oauth2_refresh_token_session_subject_fkey;

ALTER TABLE oauth2_consent_session
	ADD CONSTRAINT oauth2_consent_session_subject_fkey
		FOREIGN KEY (subject)
			REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT,
	ADD CONSTRAINT oauth2_consent_session_preconfiguration_fkey
		FOREIGN KEY (preconfiguration)
			REFERENCES oauth2_consent_preconfiguration (id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE oauth2_consent_preconfiguration
	ADD CONSTRAINT oauth2_consent_preconfiguration_subject_fkey
		FOREIGN KEY (subject)
			REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE oauth2_access_token_session
	ADD CONSTRAINT oauth2_access_token_session_challenge_id_fkey
		FOREIGN KEY (challenge_id)
			REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
	ADD CONSTRAINT oauth2_access_token_session_subject_fkey
		FOREIGN KEY (subject)
			REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE oauth2_authorization_code_session
	ADD CONSTRAINT oauth2_authorization_code_session_challenge_id_fkey
		FOREIGN KEY (challenge_id)
			REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
	ADD CONSTRAINT oauth2_authorization_code_session_subject_fkey
		FOREIGN KEY (subject)
			REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE oauth2_openid_connect_session
	ADD CONSTRAINT oauth2_openid_connect_session_challenge_id_fkey
		FOREIGN KEY (challenge_id)
			REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
	ADD CONSTRAINT oauth2_openid_connect_session_subject_fkey
		FOREIGN KEY (subject)
			REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE oauth2_pkce_request_session
	ADD CONSTRAINT oauth2_pkce_request_session_challenge_id_fkey
		FOREIGN KEY (challenge_id)
			REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
	ADD CONSTRAINT oauth2_pkce_request_session_subject_fkey
		FOREIGN KEY (subject)
			REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE oauth2_refresh_token_session
	ADD CONSTRAINT oauth2_refresh_token_session_challenge_id_fkey
		FOREIGN KEY (challenge_id)
			REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
	ADD CONSTRAINT oauth2_refresh_token_session_subject_fkey
		FOREIGN KEY (subject)
			REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;

