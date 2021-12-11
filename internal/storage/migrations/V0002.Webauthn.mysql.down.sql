CREATE TABLE IF NOT EXISTS u2f_devices (
    id INTEGER AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BLOB NOT NULL,
    public_key BLOB NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (username, description)
);

INSERT INTO u2f_devices (id, username, description, key_handle, public_key)
SELECT id, username, description, UNHEX(UPPER(kid)), public_key
FROM webauthn_devices
WHERE attestation_type = 'fido-u2f';

DROP TABLE IF EXISTS webauthn_devices;

UPDATE user_preferences
SET second_factor_method = 'u2f'
WHERE second_factor_method = 'webauthn';
