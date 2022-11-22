ALTER TABLE totp_configurations
    DROP CONSTRAINT IF EXISTS totp_configurations_pkey,
    DROP CONSTRAINT IF EXISTS totp_configurations_pkey1;

ALTER TABLE totp_configurations
    RENAME TO _bkp_DOWN_V0002_totp_configurations;

ALTER TABLE webauthn_devices
    DROP CONSTRAINT IF EXISTS webauthn_devices_pkey,
    DROP CONSTRAINT IF EXISTS webauthn_devices_pkey1;

ALTER TABLE webauthn_devices
    RENAME TO _bkp_DOWN_V0002_webauthn_devices;

CREATE TABLE IF NOT EXISTS totp_configurations (
    id SERIAL CONSTRAINT totp_configurations_pkey PRIMARY KEY,
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
FROM _bkp_DOWN_V0002_totp_configurations;

CREATE TABLE IF NOT EXISTS u2f_devices (
    id SERIAL CONSTRAINT u2f_devices_pkey PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BYTEA NOT NULL,
    public_key BYTEA NOT NULL
);

CREATE UNIQUE INDEX u2f_devices_lookup_key ON u2f_devices (username, description);

INSERT INTO u2f_devices (id, username, description, key_handle, public_key)
SELECT id, username, description, DECODE(kid, 'base64'), public_key
FROM _bkp_DOWN_V0002_webauthn_devices
WHERE attestation_type = 'fido-u2f';

UPDATE user_preferences
SET second_factor_method = 'u2f'
WHERE second_factor_method = 'webauthn';

DROP TABLE IF EXISTS _bkp_DOWN_V0002_totp_configurations;
DROP TABLE IF EXISTS _bkp_DOWN_V0002_webauthn_devices;
DROP TABLE IF EXISTS _bkp_UP_V0002_totp_configurations;
DROP TABLE IF EXISTS _bkp_UP_V0002_u2f_devices;
