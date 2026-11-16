DROP INDEX IF EXISTS oauth2_device_code_session_challenge_id_idx;

ALTER TABLE oauth2_consent_session
    DROP COLUMN binding;
