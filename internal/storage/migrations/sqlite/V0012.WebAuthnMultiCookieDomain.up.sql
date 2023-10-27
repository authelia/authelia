CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME NULL DEFAULT NULL,
    rpid VARCHAR(512) NOT NULL,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(64) NOT NULL,
    kid VARCHAR(512) NOT NULL,
    aaguid CHAR(36) NULL,
    attestation_type VARCHAR(32),
    attachment VARCHAR(64),
    transport VARCHAR(64) DEFAULT '',
    sign_count INTEGER DEFAULT 0,
    clone_warning BOOLEAN NOT NULL DEFAULT FALSE,
    legacy BOOLEAN NOT NULL DEFAULT FALSE,
    discoverable BOOLEAN NOT NULL,
    present BOOLEAN NOT NULL DEFAULT FALSE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    backup_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    backup_state BOOLEAN NOT NULL DEFAULT FALSE,
    public_key BLOB NOT NULL
);

CREATE UNIQUE INDEX webauthn_credentials_kid_key ON webauthn_credentials (kid);
CREATE UNIQUE INDEX webauthn_credentials_lookup_key ON webauthn_credentials (rpid, username, description);

INSERT INTO webauthn_credentials (created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key)
SELECT created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, 'cross-platform', transport, sign_count, clone_warning, TRUE, FALSE, FALSE, FALSE, FALSE, FALSE, public_key
FROM webauthn_devices;

DROP TABLE IF EXISTS webauthn_devices;

CREATE TABLE IF NOT EXISTS webauthn_users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    rpid VARCHAR(512) NOT NULL,
    username VARCHAR(100) NOT NULL,
    userid CHAR(64) NOT NULL
);

CREATE UNIQUE INDEX webauthn_users_lookup_key ON webauthn_users (rpid, username);
