CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_session', 'oauth2_consent_session_subject_fkey');
CALL PROC_DROP_FOREIGN_KEY('oauth2_consent_session', 'oauth2_consent_session_preconfiguration_fkey');
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

CALL PROC_DROP_INDEX('duo_devices', 'duo_devices_username_key');
CALL PROC_DROP_INDEX('encryption', 'encryption_name_key');
CALL PROC_DROP_INDEX('identity_verification', 'identity_verification_jti_key');
CALL PROC_DROP_INDEX('totp_configurations', 'totp_configurations_username_key');
CALL PROC_DROP_INDEX('user_opaque_identifier', 'user_opaque_identifier_lookup_key');
CALL PROC_DROP_INDEX('user_opaque_identifier', 'user_opaque_identifier_identifier_key');
CALL PROC_DROP_INDEX('user_preferences', 'user_preferences_username_key');
CALL PROC_DROP_INDEX('webauthn_devices', 'webauthn_devices_kid_key');
CALL PROC_DROP_INDEX('webauthn_devices', 'webauthn_devices_lookup_key');

CREATE UNIQUE INDEX username ON duo_devices (username);
CREATE UNIQUE INDEX name ON encryption (name);
CREATE UNIQUE INDEX jti ON identity_verification (jti);
CREATE UNIQUE INDEX username ON totp_configurations (username);
CREATE UNIQUE INDEX user_opaque_identifier_identifier_key ON user_opaque_identifier (identifier);
CREATE UNIQUE INDEX user_opaque_identifier_service_sector_id_username_key ON user_opaque_identifier (service, sector_id, username);
CREATE UNIQUE INDEX username ON user_preferences (username);
CREATE UNIQUE INDEX kid ON webauthn_devices (kid);
CREATE UNIQUE INDEX username ON webauthn_devices (username, description);

UPDATE webauthn_devices
SET aaguid = '00000000-00000000-00000000-00000000'
WHERE aaguid IS NULL;

ALTER TABLE webauthn_devices
    MODIFY aaguid CHAR(36) NOT NULL;

ALTER TABLE oauth2_consent_session
    ADD CONSTRAINT oauth2_consent_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT,
    ADD CONSTRAINT oauth2_consent_session_preconfiguration_fkey
        FOREIGN KEY (preconfiguration)
            REFERENCES oauth2_consent_preconfiguration (id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE oauth2_consent_preconfiguration
    ADD CONSTRAINT oauth2_consent_preconfiguration_subjct_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE oauth2_access_token_session
    ADD CONSTRAINT oauth2_access_token_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_access_token_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE oauth2_authorization_code_session
    ADD CONSTRAINT oauth2_authorization_code_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_authorization_code_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE oauth2_openid_connect_session
    ADD CONSTRAINT oauth2_openid_connect_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_openid_connect_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE oauth2_pkce_request_session
    ADD CONSTRAINT oauth2_pkce_request_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_pkce_request_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE oauth2_refresh_token_session
    ADD CONSTRAINT oauth2_refresh_token_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_refresh_token_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

DROP PROCEDURE IF EXISTS PROC_DROP_FOREIGN_KEY;
DROP PROCEDURE IF EXISTS PROC_DROP_INDEX;
