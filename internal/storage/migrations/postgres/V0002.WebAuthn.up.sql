ALTER TABLE totp_configurations
    DROP CONSTRAINT IF EXISTS totp_configurations_pkey,
    DROP CONSTRAINT IF EXISTS totp_configurations_pkey1;

ALTER TABLE totp_configurations
    RENAME TO _bkp_UP_V0002_totp_configurations;

ALTER TABLE u2f_devices
    RENAME TO _bkp_UP_V0002_u2f_devices;

DROP INDEX IF EXISTS totp_configurations_username_key;

CREATE TABLE IF NOT EXISTS totp_configurations (
    id SERIAL CONSTRAINT totp_configurations_pkey PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    secret BYTEA NOT NULL
);

CREATE UNIQUE INDEX totp_configurations_username_key ON totp_configurations (username);

INSERT INTO totp_configurations (id, username, issuer, algorithm, digits, period, secret)
SELECT id, username, issuer, algorithm, digits, period, secret
FROM _bkp_UP_V0002_totp_configurations;

CREATE TABLE IF NOT EXISTS webauthn_devices (
    id SERIAL CONSTRAINT webauthn_devices_pkey PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    rpid TEXT,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    kid VARCHAR(512) NOT NULL,
    public_key BYTEA NOT NULL,
    attestation_type VARCHAR(32),
    transport VARCHAR(64) DEFAULT '',
    aaguid CHAR(36) NOT NULL,
    sign_count INTEGER DEFAULT 0,
    clone_warning BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX webauthn_devices_kid_key ON webauthn_devices (kid);
CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (username, description);

INSERT INTO webauthn_devices (id, rpid, username, description, kid, public_key, attestation_type, aaguid, sign_count)
SELECT id, '', username, description, ENCODE(key_handle::BYTEA, 'base64'), public_key, 'fido-u2f', '00000000-0000-0000-0000-000000000000', 0
FROM _bkp_UP_V0002_u2f_devices;

UPDATE user_preferences
SET second_factor_method = 'webauthn'
WHERE second_factor_method = 'u2f';
