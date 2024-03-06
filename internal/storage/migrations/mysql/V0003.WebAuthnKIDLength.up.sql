ALTER TABLE webauthn_devices
    RENAME _bkp_UP_V0003_webauthn_devices;

CREATE TABLE IF NOT EXISTS webauthn_devices (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
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
    UNIQUE KEY (username, description),
    UNIQUE KEY (kid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

INSERT INTO webauthn_devices (id, created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning)
SELECT id, created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning
FROM _bkp_UP_V0003_webauthn_devices;

DROP TABLE IF EXISTS _bkp_UP_V0003_webauthn_devices;
