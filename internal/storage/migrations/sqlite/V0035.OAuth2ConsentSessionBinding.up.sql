ALTER TABLE oauth2_consent_session
    ADD COLUMN binding VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX oauth2_device_code_session_challenge_id_idx ON oauth2_device_code_session (challenge_id);
