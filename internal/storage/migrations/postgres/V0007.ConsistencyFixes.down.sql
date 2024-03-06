UPDATE webauthn_devices
SET aaguid = '00000000-00000000-00000000-00000000'
WHERE aaguid IS NULL;

ALTER TABLE webauthn_devices
    ALTER COLUMN aaguid SET NOT NULL;

ALTER TABLE totp_configurations
    DROP CONSTRAINT IF EXISTS totp_configurations_username_key1,
    DROP CONSTRAINT IF EXISTS totp_configurations_username_key;

ALTER TABLE webauthn_devices
    DROP CONSTRAINT IF EXISTS webauthn_devices_username_description_key1,
    DROP CONSTRAINT IF EXISTS webauthn_devices_kid_key1,
    DROP CONSTRAINT IF EXISTS webauthn_devices_lookup_key1,
    DROP CONSTRAINT IF EXISTS webauthn_devices_username_description_key,
    DROP CONSTRAINT IF EXISTS webauthn_devices_kid_key,
    DROP CONSTRAINT IF EXISTS webauthn_devices_lookup_key;

DROP INDEX IF EXISTS totp_configurations_username_key1;
DROP INDEX IF EXISTS webauthn_devices_username_description_key1;
DROP INDEX IF EXISTS webauthn_devices_kid_key1;
DROP INDEX IF EXISTS webauthn_devices_lookup_key1;
DROP INDEX IF EXISTS totp_configurations_username_key;
DROP INDEX IF EXISTS webauthn_devices_username_description_key;
DROP INDEX IF EXISTS webauthn_devices_kid_key;
DROP INDEX IF EXISTS webauthn_devices_lookup_key;

CREATE UNIQUE INDEX totp_configurations_username_key1 ON totp_configurations (username);
CREATE UNIQUE INDEX webauthn_devices_kid_key1 ON webauthn_devices (kid);
CREATE UNIQUE INDEX webauthn_devices_lookup_key1 ON webauthn_devices (username, description);

ALTER TABLE oauth2_consent_session
    DROP CONSTRAINT oauth2_consent_session_subject_fkey,
    DROP CONSTRAINT oauth2_consent_session_preconfiguration_fkey;

ALTER TABLE oauth2_consent_preconfiguration
    DROP CONSTRAINT oauth2_consent_preconfiguration_subject_fkey;

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
