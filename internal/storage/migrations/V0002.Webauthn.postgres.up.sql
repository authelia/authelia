CREATE TABLE IF NOT EXISTS webauthn_devices (
    id SERIAL,
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

INSERT INTO webauthn_devices (id, username, description, kid, public_key, attestation_type, aaguid, sign_count, clone_warning)
SELECT id, username, description, ENCODE(key_handle::BYTEA, 'base64'), public_key, 'fido-u2f', '00000000-0000-0000-0000-000000000000', 0, FALSE
FROM u2f_devices;

UPDATE user_preferences
SET second_factor_method = 'webauthn'
WHERE second_factor_method = 'u2f';

DROP TABLE IF EXISTS u2f_devices;
