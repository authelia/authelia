ALTER TABLE oauth2_blacklisted_jti
    ADD COLUMN client_id VARCHAR(255) NOT NULL DEFAULT '' AFTER id;

DROP INDEX oauth2_blacklisted_jti_signature_key ON oauth2_blacklisted_jti;

CREATE UNIQUE INDEX oauth2_blacklisted_jti_client_id_signature_key ON oauth2_blacklisted_jti (client_id, signature);

CREATE TABLE IF NOT EXISTS oauth2_request_replay (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    type VARCHAR(32) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX oauth2_request_replay_type_client_id_signature_key ON oauth2_request_replay (type, client_id, signature);
CREATE INDEX oauth2_request_replay_expires_at_idx ON oauth2_request_replay (expires_at);

DELETE older FROM oauth2_device_code_session older
    INNER JOIN oauth2_device_code_session newer
        ON newer.signature = older.signature AND newer.id > older.id;

DELETE older FROM oauth2_device_code_session older
    INNER JOIN oauth2_device_code_session newer
        ON newer.user_code_signature = older.user_code_signature AND newer.id > older.id;

CREATE UNIQUE INDEX oauth2_device_code_session_signature_key ON oauth2_device_code_session (signature);
CREATE UNIQUE INDEX oauth2_device_code_session_user_code_signature_key ON oauth2_device_code_session (user_code_signature);
