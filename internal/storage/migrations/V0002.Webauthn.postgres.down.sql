ALTER TABLE totp_configurations RENAME TO _bkp_DOWN_V0002_totp_configurations;
ALTER TABLE webauthn_devices RENAME TO _bkp_DOWN_V0002_webauthn_devices;

CREATE TABLE totp_configurations (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    secret BYTEA NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

INSERT INTO totp_configurations (id, username, issuer, algorithm, digits, period, secret)
SELECT id, username, issuer, algorithm, digits, period, secret
FROM _bkp_DOWN_V0002_totp_configurations;

CREATE TABLE u2f_devices (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BYTEA NOT NULL,
    public_key BYTEA NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username, description)
);

INSERT INTO u2f_devices (id, username, description, key_handle, public_key)
SELECT id, username, description, DECODE(kid, 'base64'), public_key
FROM _bkp_DOWN_V0002_webauthn_devices
WHERE attestation_type = 'fido-u2f';

UPDATE user_preferences
SET second_factor_method = 'u2f'
WHERE second_factor_method = 'webauthn';

DROP TABLE IF EXISTS _bkp_DOWN_V0002_totp_configurations;
DROP TABLE IF EXISTS _bkp_DOWN_V0002_webauthn_devices;
