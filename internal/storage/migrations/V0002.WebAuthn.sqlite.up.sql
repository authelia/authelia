ALTER TABLE totp_configurations
    RENAME TO _bkp_UP_V0002_totp_configurations;

ALTER TABLE u2f_devices
    RENAME TO _bkp_UP_V0002_u2f_devices;

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

INSERT INTO totp_configurations (id, username, issuer, algorithm, digits, period, secret)
SELECT id, username, issuer, algorithm, digits, period, secret
FROM _bkp_UP_V0002_totp_configurations;

DROP TABLE IF EXISTS _bkp_UP_V0002_totp_configurations;

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

INSERT INTO webauthn_devices (id, rpid, username, description, kid, public_key, attestation_type, aaguid, sign_count)
SELECT id, '', username, description, BIN2B64(key_handle), public_key, 'fido-u2f', '00000000-0000-0000-0000-000000000000', 0
FROM _bkp_UP_V0002_u2f_devices;

DROP TABLE IF EXISTS _bkp_UP_V0002_u2f_devices;

UPDATE user_preferences
SET second_factor_method = 'webauthn'
WHERE second_factor_method = 'u2f';
