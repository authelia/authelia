PRAGMA foreign_keys=off;

DROP INDEX IF EXISTS oauth2_access_token_session_request_id_idx;
DROP INDEX IF EXISTS oauth2_access_token_session_client_id_idx;
DROP INDEX IF EXISTS oauth2_access_token_session_client_id_subject_idx;

CREATE TABLE IF NOT EXISTS _bkp_UP_V0011_oauth2_access_token_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NULL DEFAULT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NULL DEFAULT NULL,
    requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL
);

INSERT INTO _bkp_UP_V0011_oauth2_access_token_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM oauth2_access_token_session
ORDER BY id;

DROP TABLE IF EXISTS oauth2_access_token_session;

CREATE TABLE IF NOT EXISTS oauth2_access_token_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NULL DEFAULT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(768) NOT NULL,
    subject CHAR(36) NULL DEFAULT NULL,
    requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL,
    CONSTRAINT oauth2_access_token_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT oauth2_access_token_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX oauth2_access_token_session_request_id_idx ON oauth2_access_token_session (request_id);
CREATE INDEX oauth2_access_token_session_client_id_idx ON oauth2_access_token_session (client_id);
CREATE INDEX oauth2_access_token_session_client_id_subject_idx ON oauth2_access_token_session (client_id, subject);

INSERT INTO oauth2_access_token_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM _bkp_UP_V0011_oauth2_access_token_session
ORDER BY id;

DROP TABLE IF EXISTS _bkp_UP_V0011_oauth2_access_token_session;

PRAGMA foreign_keys=on;
