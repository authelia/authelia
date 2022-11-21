DROP TABLE IF EXISTS oauth2_authorization_code_session;
DROP TABLE IF EXISTS oauth2_access_token_session;
DROP TABLE IF EXISTS oauth2_refresh_token_session;
DROP TABLE IF EXISTS oauth2_pkce_request_session;
DROP TABLE IF EXISTS oauth2_openid_connect_session;
DROP TABLE IF EXISTS oauth2_consent_session;
DROP TABLE IF EXISTS oauth2_consent_preconfiguration;

CREATE TABLE IF NOT EXISTS oauth2_consent_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    subject CHAR(36) NULL DEFAULT NULL,
    authorized BOOLEAN NOT NULL DEFAULT FALSE,
    granted BOOLEAN NOT NULL DEFAULT FALSE,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP NULL DEFAULT NULL,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    form_data TEXT NOT NULL,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    CONSTRAINT oauth2_consent_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE UNIQUE INDEX oauth2_consent_session_challenge_id_key ON oauth2_consent_session (challenge_id);

CREATE TABLE IF NOT EXISTS oauth2_authorization_code_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL,
    CONSTRAINT oauth2_authorization_code_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT oauth2_authorization_code_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE INDEX oauth2_authorization_code_session_request_id_idx ON oauth2_authorization_code_session (request_id);
CREATE INDEX oauth2_authorization_code_session_client_id_idx ON oauth2_authorization_code_session (client_id);
CREATE INDEX oauth2_authorization_code_session_client_id_subject_idx ON oauth2_authorization_code_session (client_id, subject);

CREATE TABLE IF NOT EXISTS oauth2_access_token_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE INDEX oauth2_access_token_session_request_id_idx ON oauth2_access_token_session (request_id);
CREATE INDEX oauth2_access_token_session_client_id_idx ON oauth2_access_token_session (client_id);
CREATE INDEX oauth2_access_token_session_client_id_subject_idx ON oauth2_access_token_session (client_id, subject);

CREATE TABLE IF NOT EXISTS oauth2_refresh_token_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL,
    CONSTRAINT oauth2_refresh_token_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT oauth2_refresh_token_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE INDEX oauth2_refresh_token_session_request_id_idx ON oauth2_refresh_token_session (request_id);
CREATE INDEX oauth2_refresh_token_session_client_id_idx ON oauth2_refresh_token_session (client_id);
CREATE INDEX oauth2_refresh_token_session_client_id_subject_idx ON oauth2_refresh_token_session (client_id, subject);

CREATE TABLE IF NOT EXISTS oauth2_pkce_request_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL,
    CONSTRAINT oauth2_pkce_request_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT oauth2_pkce_request_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE INDEX oauth2_pkce_request_session_request_id_idx ON oauth2_pkce_request_session (request_id);
CREATE INDEX oauth2_pkce_request_session_client_id_idx ON oauth2_pkce_request_session (client_id);
CREATE INDEX oauth2_pkce_request_session_client_id_subject_idx ON oauth2_pkce_request_session (client_id, subject);

CREATE TABLE IF NOT EXISTS oauth2_openid_connect_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL,
    CONSTRAINT oauth2_openid_connect_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT oauth2_openid_connect_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE INDEX oauth2_openid_connect_session_request_id_idx ON oauth2_openid_connect_session (request_id);
CREATE INDEX oauth2_openid_connect_session_client_id_idx ON oauth2_openid_connect_session (client_id);
CREATE INDEX oauth2_openid_connect_session_client_id_subject_idx ON oauth2_openid_connect_session (client_id, subject);
