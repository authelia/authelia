CREATE TABLE IF NOT EXISTS webauthn_devices (
    id SERIAL,
    ip VARCHAR(39) NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    rpid TEXT,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    kid VARCHAR(100) NOT NULL,
    public_key BYTEA NOT NULL,
    attestation_type VARCHAR(32),
    transport VARCHAR(20) DEFAULT '',
    aaguid CHAR(36) NOT NULL,
    sign_count INTEGER DEFAULT 0,
    clone_warning BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    UNIQUE (username, description),
    UNIQUE (kid)
);

INSERT INTO webauthn_devices (id, ip, username, description, kid, public_key, attestation_type, aaguid, sign_count, clone_warning)
SELECT id, '0.0.0.0', username, description, ENCODE(key_handle::BYTEA, 'base64'), public_key, 'fido-u2f', '00000000-0000-0000-0000-000000000000', 0, FALSE
FROM u2f_devices;

UPDATE user_preferences
SET second_factor_method = 'webauthn'
WHERE second_factor_method = 'u2f';

DROP TABLE IF EXISTS u2f_devices;


ALTER TABLE totp_configurations RENAME TO _bkp_UP_V0002_totp_configurations;

CREATE TABLE IF NOT EXISTS totp_configurations (
    id SERIAL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip VARCHAR(39) NOT NULL,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    secret BYTEA NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

INSERT INTO totp_configurations (id, ip, username, issuer, algorithm, digits, period, secret)
SELECT id, '0.0.0.0', username, issuer, algorithm, digits, period, secret
FROM _bkp_UP_V0002_totp_configurations;

DROP TABLE IF EXISTS _bkp_UP_V0002_totp_configurations;
