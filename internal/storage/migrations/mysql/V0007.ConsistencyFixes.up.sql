DROP PROCEDURE IF EXISTS PROC_DROP_FOREIGN_KEY;
DROP PROCEDURE IF EXISTS PROC_DROP_INDEX;

CREATE PROCEDURE PROC_DROP_FOREIGN_KEY(IN tableName VARCHAR(64), IN constraintName VARCHAR(64))
BEGIN
    IF EXISTS(
        SELECT * FROM information_schema.TABLE_CONSTRAINTS
        WHERE
            TABLE_SCHEMA    = DATABASE()     AND
            TABLE_NAME      = tableName      AND
            CONSTRAINT_NAME = constraintName AND
            CONSTRAINT_TYPE = 'FOREIGN KEY')
    THEN
        SET @query = CONCAT('ALTER TABLE ', tableName, ' DROP FOREIGN KEY ', constraintName, ';');
        PREPARE stmt FROM @query;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END;

CREATE PROCEDURE PROC_DROP_INDEX(IN tableName VARCHAR(64), IN indexName VARCHAR(64))
BEGIN
    IF EXISTS(
        SELECT * FROM information_schema.STATISTICS
        WHERE
            TABLE_SCHEMA    = DATABASE()     AND
            INDEX_SCHEMA    = DATABASE()     AND
            TABLE_NAME      = tableName      AND
            INDEX_NAME      = indexName)
    THEN
        SET @query = CONCAT('ALTER TABLE ', tableName, ' DROP INDEX ', indexName, ';');
        PREPARE stmt FROM @query;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END;

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

CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_session', 'oauth2_consent_session_subject_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_session', 'oauth2_consent_session_preconfiguration_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_preconfiguration', 'oauth2_consent_preconfiguration_subjct_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_preconfiguration', 'oauth2_consent_preconfiguration_subject_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_access_token_session', 'oauth2_access_token_session_challenge_id_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_access_token_session', 'oauth2_access_token_session_subject_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_authorization_code_session', 'oauth2_authorization_code_session_challenge_id_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_authorization_code_session', 'oauth2_authorization_code_session_subject_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_openid_connect_session', 'oauth2_openid_connect_session_challenge_id_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_openid_connect_session', 'oauth2_openid_connect_session_subject_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_pkce_request_session', 'oauth2_pkce_request_session_challenge_id_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_pkce_request_session', 'oauth2_pkce_request_session_subject_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_refresh_token_session', 'oauth2_refresh_token_session_challenge_id_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_refresh_token_session', 'oauth2_refresh_token_session_subject_fkey');

CALL PROC_DROP_INDEX('duo_devices', 'username');
CALL PROC_DROP_INDEX('encryption', 'name');
CALL PROC_DROP_INDEX('identity_verification', 'jti');
CALL PROC_DROP_INDEX('totp_configurations', 'username');
CALL PROC_DROP_INDEX('user_opaque_identifier', 'user_opaque_identifier_identifier_key');
CALL PROC_DROP_INDEX('user_opaque_identifier', 'user_opaque_identifier_service_sector_id_username_key');
CALL PROC_DROP_INDEX('user_preferences', 'username');
CALL PROC_DROP_INDEX('webauthn_devices', 'username');
CALL PROC_DROP_INDEX('webauthn_devices', 'kid');

CREATE UNIQUE INDEX duo_devices_username_key ON duo_devices (username);
CREATE UNIQUE INDEX encryption_name_key ON encryption (name);
CREATE UNIQUE INDEX identity_verification_jti_key ON identity_verification (jti);
CREATE UNIQUE INDEX totp_configurations_username_key ON totp_configurations (username);
CREATE UNIQUE INDEX user_opaque_identifier_identifier_key ON user_opaque_identifier (identifier);
CREATE UNIQUE INDEX user_opaque_identifier_lookup_key ON user_opaque_identifier (service, sector_id, username);
CREATE UNIQUE INDEX user_preferences_username_key ON user_preferences (username);
CREATE UNIQUE INDEX webauthn_devices_kid_key ON webauthn_devices (kid);
CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (username, description);

ALTER TABLE webauthn_devices
    MODIFY aaguid CHAR(36) NULL;

UPDATE webauthn_devices
SET aaguid = NULL
WHERE aaguid = '' OR aaguid = '00000000-00000000-00000000-00000000';

ALTER TABLE authentication_logs
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE duo_devices
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE encryption
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE identity_verification
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE migrations
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_blacklisted_jti
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_consent_session
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_consent_preconfiguration
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_access_token_session
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_authorization_code_session
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_openid_connect_session
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_pkce_request_session
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_refresh_token_session
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE totp_configurations
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE user_opaque_identifier
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE user_preferences
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE webauthn_devices
    ENGINE=InnoDB,
    CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

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
