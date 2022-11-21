CREATE TABLE IF NOT EXISTS user_opaque_identifier (
    id SERIAL CONSTRAINT user_opaque_identifier_pkey PRIMARY KEY,
    service VARCHAR(20) NOT NULL,
    sector_id VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    identifier CHAR(36) NOT NULL
);

CREATE UNIQUE INDEX user_opaque_identifier_service_sector_id_username_key ON user_opaque_identifier (service, sector_id, username);
CREATE UNIQUE INDEX user_opaque_identifier_identifier_key ON user_opaque_identifier (identifier);

CREATE TABLE IF NOT EXISTS oauth2_blacklisted_jti (
    id SERIAL CONSTRAINT oauth2_blacklisted_jti_pkey PRIMARY KEY,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX oauth2_blacklisted_jti_signature_key ON oauth2_blacklisted_jti (signature);

CREATE TABLE IF NOT EXISTS oauth2_consent_session (
    id SERIAL CONSTRAINT oauth2_consent_session_pkey PRIMARY KEY,
    challenge_id CHAR(36) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    authorized BOOLEAN NOT NULL DEFAULT FALSE,
    granted BOOLEAN NOT NULL DEFAULT FALSE,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    form_data TEXT NOT NULL,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT ''
);

CREATE UNIQUE INDEX oauth2_consent_session_challenge_id_key ON oauth2_consent_session (challenge_id);

ALTER TABLE oauth2_consent_session
    ADD CONSTRAINT oauth2_consent_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS oauth2_authorization_code_session (
    id SERIAL CONSTRAINT oauth2_authorization_code_session_pkey PRIMARY KEY,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36),
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL
);

CREATE INDEX oauth2_authorization_code_session_request_id_idx ON oauth2_authorization_code_session (request_id);
CREATE INDEX oauth2_authorization_code_session_client_id_idx ON oauth2_authorization_code_session (client_id);
CREATE INDEX oauth2_authorization_code_session_client_id_subject_idx ON oauth2_authorization_code_session (client_id, subject);

ALTER TABLE oauth2_authorization_code_session
    ADD CONSTRAINT oauth2_authorization_code_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_authorization_code_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS oauth2_access_token_session (
    id SERIAL CONSTRAINT oauth2_access_token_session_pkey PRIMARY KEY,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL
);

CREATE INDEX oauth2_access_token_session_request_id_idx ON oauth2_access_token_session (request_id);
CREATE INDEX oauth2_access_token_session_client_id_idx ON oauth2_access_token_session (client_id);
CREATE INDEX oauth2_access_token_session_client_id_subject_idx ON oauth2_access_token_session (client_id, subject);

ALTER TABLE oauth2_access_token_session
    ADD CONSTRAINT oauth2_access_token_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_access_token_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS oauth2_refresh_token_session (
    id SERIAL CONSTRAINT oauth2_refresh_token_session_pkey PRIMARY KEY,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL
);

CREATE INDEX oauth2_refresh_token_session_request_id_idx ON oauth2_refresh_token_session (request_id);
CREATE INDEX oauth2_refresh_token_session_client_id_idx ON oauth2_refresh_token_session (client_id);
CREATE INDEX oauth2_refresh_token_session_client_id_subject_idx ON oauth2_refresh_token_session (client_id, subject);

ALTER TABLE oauth2_refresh_token_session
    ADD CONSTRAINT oauth2_refresh_token_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_refresh_token_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS oauth2_pkce_request_session (
    id SERIAL CONSTRAINT oauth2_pkce_request_session_pkey PRIMARY KEY,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL
);

CREATE INDEX oauth2_pkce_request_session_request_id_idx ON oauth2_pkce_request_session (request_id);
CREATE INDEX oauth2_pkce_request_session_client_id_idx ON oauth2_pkce_request_session (client_id);
CREATE INDEX oauth2_pkce_request_session_client_id_subject_idx ON oauth2_pkce_request_session (client_id, subject);

ALTER TABLE oauth2_pkce_request_session
    ADD CONSTRAINT oauth2_pkce_request_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_pkce_request_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS oauth2_openid_connect_session (
    id SERIAL CONSTRAINT oauth2_openid_connect_session_pkey PRIMARY KEY,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL
);

CREATE INDEX oauth2_openid_connect_session_request_id_idx ON oauth2_openid_connect_session (request_id);
CREATE INDEX oauth2_openid_connect_session_client_id_idx ON oauth2_openid_connect_session (client_id);
CREATE INDEX oauth2_openid_connect_session_client_id_subject_idx ON oauth2_openid_connect_session (client_id, subject);

ALTER TABLE oauth2_openid_connect_session
    ADD CONSTRAINT oauth2_openid_connect_session_challenge_id_fkey
        FOREIGN KEY (challenge_id)
            REFERENCES oauth2_consent_session (challenge_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT oauth2_openid_connect_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;
