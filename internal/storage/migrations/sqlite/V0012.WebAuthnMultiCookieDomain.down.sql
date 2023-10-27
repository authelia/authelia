CREATE TABLE IF NOT EXISTS webauthn_devices (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME NULL DEFAULT NULL,
    rpid TEXT,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    kid VARCHAR(512) NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type VARCHAR(32),
    transport VARCHAR(64) DEFAULT '',
    aaguid CHAR(36) NULL,
    sign_count INTEGER DEFAULT 0,
    clone_warning BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (username, description);
CREATE UNIQUE INDEX webauthn_devices_kid_key ON webauthn_devices (kid);

INSERT INTO webauthn_devices (created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning)
SELECT created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning
FROM webauthn_credentials
WHERE legacy = TRUE;

DROP TABLE IF EXISTS webauthn_credentials;
DROP TABLE IF EXISTS webauthn_users;
