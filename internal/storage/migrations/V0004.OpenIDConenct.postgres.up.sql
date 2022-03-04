CREATE TABLE IF NOT EXISTS oauth2_blacklisted_jti (
    id SERIAL,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE (signature)
);

CREATE TABLE IF NOT EXISTS oauth2_subjects (
    id SERIAL,
    sector_id VARCHAR(255) NULL DEFAULT NULL,
    subject_id VARCHAR(255) NOT NULL,
    salt CHAR(32) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE oauth2_subjects_sector_id_subject_id_idx (sector_id, subject_id)
);

CREATE TABLE IF NOT EXISTS oauth2_authorize_code_sessions (
    id SERIAL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL DEFAULT '',
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX oauth2_authorize_code_sessions_request_id_idx ON oauth2_authorize_code_sessions (request_id);
CREATE INDEX oauth2_authorize_code_sessions_client_id_idx ON oauth2_authorize_code_sessions (client_id);

CREATE TABLE IF NOT EXISTS oauth2_access_token_sessions (
    id SERIAL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL DEFAULT '',
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX oauth2_access_token_sessions_request_id_idx ON oauth2_access_token_sessions (request_id);
CREATE INDEX oauth2_access_token_sessions_client_id_idx ON oauth2_access_token_sessions (client_id);

CREATE TABLE IF NOT EXISTS oauth2_refresh_token_sessions (
    id SERIAL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL DEFAULT '',
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX oauth2_refresh_token_sessions_request_id_idx ON oauth2_refresh_token_sessions (request_id);
CREATE INDEX oauth2_refresh_token_sessions_client_id_idx ON oauth2_refresh_token_sessions (client_id);

CREATE TABLE IF NOT EXISTS oauth2_pkce_request_sessions (
    id SERIAL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL DEFAULT '',
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX oauth2_pkce_request_sessions_request_id_idx ON oauth2_pkce_request_sessions (request_id);
CREATE INDEX oauth2_pkce_request_sessions_client_id_idx ON oauth2_pkce_request_sessions (client_id);

CREATE TABLE IF NOT EXISTS oauth2_openid_connect_sessions (
    id SERIAL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL DEFAULT '',
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX oauth2_openid_connect_sessions_request_id_idx ON oauth2_openid_connect_sessions (request_id);
CREATE INDEX oauth2_openid_connect_sessions_client_id_idx ON oauth2_openid_connect_sessions (client_id);
