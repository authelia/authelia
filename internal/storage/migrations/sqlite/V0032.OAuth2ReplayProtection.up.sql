ALTER TABLE oauth2_blacklisted_jti
    ADD COLUMN client_id VARCHAR(255) NOT NULL DEFAULT '';

DROP INDEX IF EXISTS oauth2_blacklisted_jti_signature_key;

CREATE UNIQUE INDEX oauth2_blacklisted_jti_client_id_signature_key ON oauth2_blacklisted_jti (client_id, signature);

CREATE TABLE IF NOT EXISTS oauth2_request_replay (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type VARCHAR(32) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX oauth2_request_replay_type_client_id_signature_key ON oauth2_request_replay (type, client_id, signature);
CREATE INDEX oauth2_request_replay_expires_at_idx ON oauth2_request_replay (expires_at);

DELETE FROM oauth2_device_code_session
WHERE EXISTS (
    SELECT 1 FROM oauth2_device_code_session newer
    WHERE newer.signature = oauth2_device_code_session.signature AND newer.id > oauth2_device_code_session.id
);

DELETE FROM oauth2_device_code_session
WHERE EXISTS (
    SELECT 1 FROM oauth2_device_code_session newer
    WHERE newer.user_code_signature = oauth2_device_code_session.user_code_signature AND newer.id > oauth2_device_code_session.id
);

CREATE UNIQUE INDEX oauth2_device_code_session_signature_key ON oauth2_device_code_session (signature);
CREATE UNIQUE INDEX oauth2_device_code_session_user_code_signature_key ON oauth2_device_code_session (user_code_signature);
