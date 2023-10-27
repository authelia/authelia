PRAGMA foreign_keys=off;

ALTER TABLE webauthn_devices
    RENAME TO _bkp_DOWN_V0007_webauthn_devices;

CREATE TABLE IF NOT EXISTS webauthn_devices (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    rpid TEXT,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    kid VARCHAR(512) NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type VARCHAR(32),
    transport VARCHAR(64) DEFAULT '',
    aaguid CHAR(36) NOT NULL,
    sign_count INTEGER DEFAULT 0,
    clone_warning BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (username, description),
    UNIQUE (kid)
);

INSERT INTO webauthn_devices (created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning)
SELECT created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning
FROM _bkp_DOWN_V0007_webauthn_devices;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_webauthn_devices;

ALTER TABLE identity_verification
    RENAME TO _bkp_DOWN_V0007_identity_verification;

CREATE TABLE IF NOT EXISTS identity_verification (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    jti VARCHAR(36),
    iat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    issued_ip VARCHAR(39) NOT NULL,
    exp TIMESTAMP NOT NULL,
    username VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    consumed TIMESTAMP NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    UNIQUE (jti)
);

INSERT INTO identity_verification (jti, iat, issued_ip, exp, username, action, consumed, consumed_ip)
SELECT jti, iat, issued_ip, exp, username, action, consumed, consumed_ip
FROM _bkp_DOWN_V0007_identity_verification
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_identity_verification;

ALTER TABLE totp_configurations
    RENAME TO _bkp_DOWN_V0007_totp_configurations;

CREATE TABLE IF NOT EXISTS totp_configurations (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    secret BLOB NOT NULL,
    UNIQUE (username)
);

INSERT INTO totp_configurations (username, issuer, algorithm, digits, period, secret)
SELECT username, issuer, algorithm, digits, period, secret
FROM _bkp_DOWN_V0007_totp_configurations
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_totp_configurations;

ALTER TABLE duo_devices
    RENAME TO _bkp_DOWN_V0007_duo_devices;

CREATE TABLE IF NOT EXISTS duo_devices (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL,
    device VARCHAR(32) NOT NULL,
    method VARCHAR(16) NOT NULL,
    UNIQUE (username)
);

INSERT INTO duo_devices (username, device, method)
SELECT username, device, method
FROM _bkp_DOWN_V0007_duo_devices
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_duo_devices;

ALTER TABLE user_preferences
    RENAME TO _bkp_DOWN_V0007_user_preferences;

CREATE TABLE IF NOT EXISTS user_preferences (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL,
    second_factor_method VARCHAR(11) NOT NULL,
    UNIQUE (username)
);

INSERT INTO user_preferences (username, second_factor_method)
SELECT username, second_factor_method
FROM _bkp_DOWN_V0007_user_preferences
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_user_preferences;

ALTER TABLE encryption
    RENAME TO _bkp_DOWN_V0007_encryption;

CREATE TABLE IF NOT EXISTS encryption (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100),
    value BLOB NOT NULL,
    UNIQUE (name)
);

INSERT INTO encryption (name, value)
SELECT name, value
FROM _bkp_DOWN_V0007_encryption
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_encryption;

CREATE TABLE IF NOT EXISTS _bkp_DOWN_V0007_oauth2_consent_preconfiguration (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    client_id VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    scopes TEXT NOT NULL,
    audience TEXT NULL
);

INSERT INTO _bkp_DOWN_V0007_oauth2_consent_preconfiguration (client_id, subject, created_at, expires_at, revoked, scopes, audience)
SELECT client_id, subject, created_at, expires_at, revoked, scopes, audience
FROM oauth2_consent_preconfiguration
ORDER BY id;

DROP TABLE IF EXISTS oauth2_consent_preconfiguration;

CREATE TABLE IF NOT EXISTS _bkp_DOWN_V0007_oauth2_consent_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
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
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    preconfiguration INTEGER NULL DEFAULT NULL
);

INSERT INTO _bkp_DOWN_V0007_oauth2_consent_session (challenge_id, client_id, subject, authorized, granted, requested_at, responded_at, form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, preconfiguration)
SELECT challenge_id, client_id, subject, authorized, granted, requested_at, responded_at, form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, preconfiguration
FROM oauth2_consent_session
ORDER BY id;

DROP TABLE IF EXISTS oauth2_consent_session;

CREATE TABLE IF NOT EXISTS _bkp_DOWN_V0007_oauth2_authorization_code_session (
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
    session_data BLOB NOT NULL
);

INSERT INTO _bkp_DOWN_V0007_oauth2_authorization_code_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM oauth2_authorization_code_session
ORDER BY id;

DROP TABLE IF EXISTS oauth2_authorization_code_session;

CREATE TABLE IF NOT EXISTS _bkp_DOWN_V0007_oauth2_access_token_session (
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
    session_data BLOB NOT NULL
);

INSERT INTO _bkp_DOWN_V0007_oauth2_access_token_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM oauth2_access_token_session
ORDER BY id;

DROP TABLE IF EXISTS oauth2_access_token_session;

CREATE TABLE IF NOT EXISTS _bkp_DOWN_V0007_oauth2_refresh_token_session (
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
    session_data BLOB NOT NULL
);

INSERT INTO _bkp_DOWN_V0007_oauth2_refresh_token_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM oauth2_refresh_token_session
ORDER BY id;

DROP TABLE IF EXISTS oauth2_refresh_token_session;

CREATE TABLE IF NOT EXISTS _bkp_DOWN_V0007_oauth2_pkce_request_session (
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
    session_data BLOB NOT NULL
);

INSERT INTO _bkp_DOWN_V0007_oauth2_pkce_request_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM oauth2_pkce_request_session
ORDER BY id;

DROP TABLE IF EXISTS oauth2_pkce_request_session;

CREATE TABLE IF NOT EXISTS _bkp_DOWN_V0007_oauth2_openid_connect_session (
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
    session_data BLOB NOT NULL
);

INSERT INTO _bkp_DOWN_V0007_oauth2_openid_connect_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM oauth2_openid_connect_session
ORDER BY id;

DROP TABLE IF EXISTS oauth2_openid_connect_session;

DROP INDEX IF EXISTS user_opaque_identifier_identifier_key;
DROP INDEX IF EXISTS user_opaque_identifier_lookup_key;

ALTER TABLE user_opaque_identifier
    RENAME TO _bkp_DOWN_V0007_user_opaque_identifier;

CREATE TABLE IF NOT EXISTS user_opaque_identifier (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    service VARCHAR(20) NOT NULL,
    sector_id VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    identifier CHAR(36) NOT NULL
);

CREATE UNIQUE INDEX user_opaque_identifier_service_sector_id_username_key ON user_opaque_identifier (service, sector_id, username);
CREATE UNIQUE INDEX user_opaque_identifier_identifier_key ON user_opaque_identifier (identifier);

INSERT INTO user_opaque_identifier (service, sector_id, username, identifier)
SELECT service, sector_id, username, identifier
FROM _bkp_DOWN_V0007_user_opaque_identifier
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_user_opaque_identifier;

DROP INDEX IF EXISTS authentication_logs_username_idx;
DROP INDEX IF EXISTS authentication_logs_remote_ip_idx;

ALTER TABLE authentication_logs
    RENAME TO _bkp_DOWN_V0007_authentication_logs;

CREATE TABLE IF NOT EXISTS authentication_logs (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    successful BOOLEAN NOT NULL,
    banned BOOLEAN NOT NULL DEFAULT FALSE,
    username VARCHAR(100) NOT NULL,
    auth_type VARCHAR(8) NOT NULL DEFAULT '1FA',
    remote_ip VARCHAR(39) NULL DEFAULT NULL,
    request_uri TEXT,
    request_method VARCHAR(8) NOT NULL DEFAULT ''
);

CREATE INDEX authentication_logs_username_idx ON authentication_logs (time, username, auth_type);
CREATE INDEX authentication_logs_remote_ip_idx ON authentication_logs (time, remote_ip, auth_type);

INSERT INTO authentication_logs (time, successful, banned, username, auth_type, remote_ip, request_uri, request_method)
SELECT time, successful, banned, username, auth_type, remote_ip, request_uri, request_method
FROM _bkp_DOWN_V0007_authentication_logs
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_authentication_logs;

ALTER TABLE migrations
    RENAME TO _bkp_DOWN_V0007_migrations;

CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    applied TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version_before INTEGER NULL DEFAULT NULL,
    version_after INTEGER NOT NULL,
    application_version VARCHAR(128) NOT NULL
);

INSERT INTO migrations (applied, version_before, version_after, application_version)
SELECT applied, version_before, version_after, application_version
FROM _bkp_DOWN_V0007_migrations
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_migrations;

DROP INDEX IF EXISTS oauth2_blacklisted_jti_signature_key;

ALTER TABLE oauth2_blacklisted_jti
    RENAME TO _bkp_DOWN_V0007_oauth2_blacklisted_jti;

CREATE TABLE IF NOT EXISTS oauth2_blacklisted_jti (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX oauth2_blacklisted_jti_signature_key ON oauth2_blacklisted_jti (signature);

INSERT INTO oauth2_blacklisted_jti (signature, expires_at)
SELECT signature, expires_at
FROM _bkp_DOWN_V0007_oauth2_blacklisted_jti
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_blacklisted_jti;

CREATE TABLE IF NOT EXISTS oauth2_consent_preconfiguration (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    client_id VARCHAR(255) NOT NULL,
    subject CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    scopes TEXT NOT NULL,
    audience TEXT NULL,
    CONSTRAINT oauth2_consent_preconfiguration_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);

INSERT INTO oauth2_consent_preconfiguration (client_id, subject, created_at, expires_at, revoked, scopes, audience)
SELECT client_id, subject, created_at, expires_at, revoked, scopes, audience
FROM _bkp_DOWN_V0007_oauth2_consent_preconfiguration
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_consent_preconfiguration;

CREATE TABLE IF NOT EXISTS oauth2_consent_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
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
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    preconfiguration INTEGER NULL DEFAULT NULL,
    CONSTRAINT oauth2_consent_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT oauth2_consent_session_preconfiguration_fkey
        FOREIGN KEY (preconfiguration)
            REFERENCES oauth2_consent_preconfiguration (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX oauth2_consent_session_challenge_id_key ON oauth2_consent_session (challenge_id);

INSERT INTO oauth2_consent_session (challenge_id, client_id, subject, authorized, granted, requested_at, responded_at, form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, preconfiguration)
SELECT challenge_id, client_id, subject, authorized, granted, requested_at, responded_at, form_data, requested_scopes, granted_scopes, requested_audience, granted_audience, preconfiguration
FROM _bkp_DOWN_V0007_oauth2_consent_session
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_consent_session;

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

INSERT INTO oauth2_authorization_code_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM _bkp_DOWN_V0007_oauth2_authorization_code_session
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_authorization_code_session;

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

INSERT INTO oauth2_access_token_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM _bkp_DOWN_V0007_oauth2_access_token_session
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_access_token_session;

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

INSERT INTO oauth2_refresh_token_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM _bkp_DOWN_V0007_oauth2_refresh_token_session
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_refresh_token_session;

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

INSERT INTO oauth2_pkce_request_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM _bkp_DOWN_V0007_oauth2_pkce_request_session
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_pkce_request_session;

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

INSERT INTO oauth2_openid_connect_session (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data)
SELECT challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, requested_audience, granted_audience, active, revoked, form_data, session_data
FROM _bkp_DOWN_V0007_oauth2_openid_connect_session
ORDER BY id;

DROP TABLE IF EXISTS _bkp_DOWN_V0007_oauth2_openid_connect_session;

PRAGMA foreign_keys=on;
