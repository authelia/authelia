CREATE TABLE IF NOT EXISTS u2f_devices (
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
FROM webauthn_devices
WHERE attestation_type = 'fido-u2f';

UPDATE user_preferences
SET second_factor_method = 'u2f'
WHERE second_factor_method = 'webauthn';

DROP TABLE IF EXISTS webauthn_devices;
