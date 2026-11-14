CREATE TABLE IF NOT EXISTS app_authenticator_devices (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    device_id VARCHAR(256) NOT NULL,
    username VARCHAR(100) NOT NULL,
    base_url TEXT NOT NULL,
    name VARCHAR(64) NOT NULL,
    model VARCHAR(64) NOT NULL,
    platform VARCHAR(16) NOT NULL,
    approval_public_key BLOB NOT NULL,
    transport_public_key BLOB NOT NULL,
    attestation_format VARCHAR(32) NOT NULL,
    attestation_identity VARCHAR(255) NULL DEFAULT NULL,
    attestation_detail VARCHAR(32) NULL DEFAULT NULL,
    attestation_key_id BLOB NULL DEFAULT NULL,
    pull_id_current VARBINARY(255) NOT NULL,
    pull_id_previous VARBINARY(255) NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX app_authenticator_devices_device_id_key ON app_authenticator_devices (device_id);
CREATE INDEX app_authenticator_devices_username_idx ON app_authenticator_devices (username);

CREATE TABLE IF NOT EXISTS app_authenticator_registrations (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    code VARBINARY(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    portal_binding VARBINARY(255) NOT NULL,
    base_url TEXT NOT NULL,
    number VARCHAR(4) NOT NULL,
    nonce_length SMALLINT NOT NULL,
    state VARCHAR(16) NOT NULL,
    begun_at TIMESTAMP NULL DEFAULT NULL,
    session_expires_at TIMESTAMP NULL DEFAULT NULL,
    confirmation_id VARBINARY(255) NULL DEFAULT NULL,
    challenge VARCHAR(64) NULL DEFAULT NULL,
    code_challenge VARCHAR(64) NULL DEFAULT NULL,
    device_platform VARCHAR(16) NULL DEFAULT NULL,
    device_model VARCHAR(64) NULL DEFAULT NULL,
    device_name VARCHAR(64) NULL DEFAULT NULL,
    device_id VARCHAR(256) NULL DEFAULT NULL,
    CONSTRAINT app_authenticator_registrations_nonce_length_chk CHECK (nonce_length BETWEEN 2 AND 4)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX app_authenticator_registrations_code_key ON app_authenticator_registrations (code);
CREATE UNIQUE INDEX app_authenticator_registrations_confirmation_id_key ON app_authenticator_registrations (confirmation_id);
CREATE INDEX app_authenticator_registrations_portal_binding_idx ON app_authenticator_registrations (portal_binding);

CREATE TABLE IF NOT EXISTS app_authenticator_challenges (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    confirmation_id VARCHAR(256) NOT NULL,
    username VARCHAR(100) NOT NULL,
    portal_binding VARBINARY(255) NOT NULL,
    challenge VARCHAR(64) NOT NULL,
    number VARCHAR(4) NOT NULL,
    nonce_length SMALLINT NOT NULL,
    state VARCHAR(16) NOT NULL,
    bound_device_id VARCHAR(256) NULL DEFAULT NULL,
    bound_code_challenge VARCHAR(64) NULL DEFAULT NULL,
    decided_at TIMESTAMP NULL DEFAULT NULL,
    decided_device_id VARCHAR(256) NULL DEFAULT NULL,
    portal_url TEXT NULL DEFAULT NULL,
    requested_url TEXT NULL DEFAULT NULL,
    remote_ip VARCHAR(64) NULL DEFAULT NULL,
    user_agent TEXT NULL DEFAULT NULL,
    consumed_at TIMESTAMP NULL DEFAULT NULL,
    CONSTRAINT app_authenticator_challenges_nonce_length_chk CHECK (nonce_length BETWEEN 2 AND 4)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX app_authenticator_challenges_confirmation_id_key ON app_authenticator_challenges (confirmation_id);
CREATE INDEX app_authenticator_challenges_username_state_idx ON app_authenticator_challenges (username, state);
CREATE INDEX app_authenticator_challenges_portal_binding_idx ON app_authenticator_challenges (portal_binding);

CREATE TABLE IF NOT EXISTS app_authenticator_request_nonces (
    device_id VARCHAR(256) NOT NULL,
    nonce CHAR(22) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    CONSTRAINT app_authenticator_request_nonces_pkey PRIMARY KEY (device_id, nonce)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE INDEX app_authenticator_request_nonces_expires_at_idx ON app_authenticator_request_nonces (expires_at);
