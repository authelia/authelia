CREATE TABLE IF NOT EXISTS oauth2_consent_preconfiguration (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    client_id VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    scopes TEXT NOT NULL,
    audience TEXT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

ALTER TABLE oauth2_consent_preconfiguration
    ADD CONSTRAINT oauth2_consent_preconfiguration_subjct_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;

INSERT INTO oauth2_consent_preconfiguration (client_id, subject, created_at, expires_at, scopes, audience)
SELECT client_id, subject, responded_at, expires_at, granted_scopes, granted_audience
FROM oauth2_consent_session
WHERE expires_at IS NOT NULL AND responded_at IS NOT NULL
ORDER BY responded_at;

DROP TABLE IF EXISTS oauth2_access_token_session;
DROP TABLE IF EXISTS oauth2_authorization_code_session;
DROP TABLE IF EXISTS oauth2_openid_connect_session;
DROP TABLE IF EXISTS oauth2_pkce_request_session;
DROP TABLE IF EXISTS oauth2_refresh_token_session;
DROP TABLE IF EXISTS oauth2_consent_session;

CREATE TABLE IF NOT EXISTS oauth2_consent_session (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    challenge_id CHAR(36) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    authorized BOOLEAN NOT NULL DEFAULT FALSE,
    granted BOOLEAN NOT NULL DEFAULT FALSE,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP NULL DEFAULT NULL,
    form_data TEXT NOT NULL,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL,
    granted_audience TEXT NULL,
    preconfiguration INTEGER NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX oauth2_consent_session_challenge_id_key ON oauth2_consent_session (challenge_id);

ALTER TABLE oauth2_consent_session
    ADD CONSTRAINT oauth2_consent_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT,
    ADD CONSTRAINT oauth2_consent_session_preconfiguration_fkey
        FOREIGN KEY (preconfiguration)
            REFERENCES oauth2_consent_preconfiguration (id) ON UPDATE CASCADE ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS oauth2_access_token_session (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL,
    granted_audience TEXT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

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

CREATE TABLE IF NOT EXISTS oauth2_authorization_code_session (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL,
    granted_audience TEXT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

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

CREATE TABLE IF NOT EXISTS oauth2_openid_connect_session (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL,
    granted_audience TEXT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

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

CREATE TABLE IF NOT EXISTS oauth2_pkce_request_session (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL,
    granted_audience TEXT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

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

CREATE TABLE IF NOT EXISTS oauth2_refresh_token_session (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    challenge_id CHAR(36) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL,
    granted_audience TEXT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

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
