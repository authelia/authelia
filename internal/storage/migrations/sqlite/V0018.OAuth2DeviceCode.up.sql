CREATE TABLE IF NOT EXISTS oauth2_device_code_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NULL DEFAULT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    user_code_signature VARCHAR(255) NOT NULL,
    status INTEGER NOT NULL,
    subject CHAR(36) NULL DEFAULT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL,
    CONSTRAINT oauth2_device_code_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT oauth2_device_code_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX oauth2_device_code_session_request_id_idx ON oauth2_device_code_session (request_id);
CREATE INDEX oauth2_device_code_session_client_id_idx ON oauth2_device_code_session (client_id);
CREATE INDEX oauth2_device_code_session_client_id_subject_idx ON oauth2_device_code_session (client_id, subject);
