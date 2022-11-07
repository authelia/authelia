DROP PROCEDURE IF EXISTS PROC_DROP_FOREIGN_KEY;
CREATE PROCEDURE PROC_DROP_FOREIGN_KEY(IN tableName VARCHAR(64), IN constraintName VARCHAR(64))
BEGIN
	IF EXISTS(
		SELECT * FROM information_schema.table_constraints
		WHERE
			table_schema    = DATABASE()     AND
			table_name      = tableName      AND
			constraint_name = constraintName AND
			constraint_type = 'FOREIGN KEY')
	THEN
		SET @query = CONCAT('ALTER TABLE ', tableName, ' DROP FOREIGN KEY ', constraintName, ';');
		PREPARE stmt FROM @query;
		EXECUTE stmt;
		DEALLOCATE PREPARE stmt;
	END IF;
END;

DROP TABLE _bkp_UP_V0002_totp_configurations;
DROP TABLE _bkp_UP_V0002_u2f_devices;

ALTER TABLE oauth2_consent_session DROP FOREIGN KEY oauth2_consent_session_subject_fkey;
ALTER TABLE oauth2_consent_session DROP FOREIGN KEY oauth2_consent_session_preconfiguration_fkey;
CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_session', 'oauth2_consent_preconfiguration_subjct_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_preconfiguration', 'oauth2_consent_preconfiguration_subject_fkey');
ALTER TABLE oauth2_access_token_session DROP FOREIGN KEY oauth2_access_token_session_challenge_id_fkey;
ALTER TABLE oauth2_access_token_session DROP FOREIGN KEY oauth2_access_token_session_subject_fkey;
ALTER TABLE oauth2_authorization_code_session DROP FOREIGN KEY oauth2_authorization_code_session_challenge_id_fkey;
ALTER TABLE oauth2_authorization_code_session DROP FOREIGN KEY oauth2_authorization_code_session_subject_fkey;
ALTER TABLE oauth2_openid_connect_session DROP FOREIGN KEY oauth2_openid_connect_session_challenge_id_fkey;
ALTER TABLE oauth2_openid_connect_session DROP FOREIGN KEY oauth2_openid_connect_session_subject_fkey;
ALTER TABLE oauth2_pkce_request_session DROP FOREIGN KEY oauth2_pkce_request_session_challenge_id_fkey;
ALTER TABLE oauth2_pkce_request_session DROP FOREIGN KEY oauth2_pkce_request_session_subject_fkey;
ALTER TABLE oauth2_refresh_token_session DROP FOREIGN KEY oauth2_refresh_token_session_challenge_id_fkey;
ALTER TABLE oauth2_refresh_token_session DROP FOREIGN KEY oauth2_refresh_token_session_subject_fkey;

ALTER TABLE authentication_logs CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE duo_devices CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE encryption CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE identity_verification CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE migrations CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_blacklisted_jti CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_consent_session CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_consent_preconfiguration CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_access_token_session CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_authorization_code_session CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_openid_connect_session CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_pkce_request_session CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE oauth2_refresh_token_session CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE totp_configurations CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE user_opaque_identifier CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE user_preferences CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
ALTER TABLE webauthn_devices CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_consent_session ADD CONSTRAINT oauth2_consent_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_consent_session ADD CONSTRAINT oauth2_consent_session_preconfiguration_fkey FOREIGN KEY (preconfiguration) REFERENCES oauth2_consent_preconfiguration (id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE oauth2_consent_preconfiguration ADD CONSTRAINT oauth2_consent_preconfiguration_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_access_token_session ADD CONSTRAINT oauth2_access_token_session_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE oauth2_access_token_session ADD CONSTRAINT oauth2_access_token_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_authorization_code_session ADD CONSTRAINT oauth2_authorization_code_session_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE oauth2_authorization_code_session ADD CONSTRAINT oauth2_authorization_code_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_openid_connect_session ADD CONSTRAINT oauth2_openid_connect_session_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE oauth2_openid_connect_session ADD CONSTRAINT oauth2_openid_connect_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_pkce_request_session ADD CONSTRAINT oauth2_pkce_request_session_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE oauth2_pkce_request_session ADD CONSTRAINT oauth2_pkce_request_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_refresh_token_session ADD CONSTRAINT oauth2_refresh_token_session_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE oauth2_refresh_token_session ADD CONSTRAINT oauth2_refresh_token_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
