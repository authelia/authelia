CREATE TABLE IF NOT EXISTS webauthn_devices (
    id INTEGER,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    kid BLOB NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type VARCHAR(32),
    aaguid CHAR(36) NOT NULL,
    sign_count INTEGER,
    clone_warning BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    UNIQUE (username, description)
);

INSERT INTO webauthn_devices (id, username, description, kid, public_key, attestation_type, aaguid, sign_count)
SELECT id, username, description, LOWER(HEX(key_handle)), public_key, 'fido-u2f', '00000000-0000-0000-0000-000000000000', 0
FROM u2f_devices;

UPDATE user_preferences
SET second_factor_method = 'webauthn'
WHERE second_factor_method = 'u2f';

DROP TABLE IF EXISTS u2f_devices;

