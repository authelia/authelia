ALTER  TABLE webauthn_devices
    DROP CONSTRAINT IF EXISTS webauthn_devices_pkey,
    DROP CONSTRAINT IF EXISTS webauthn_devices_pkey1;

ALTER TABLE webauthn_devices
    RENAME TO _bkp_UP_V0003_webauthn_devices;

DROP INDEX IF EXISTS webauthn_devices_kid_key;
DROP INDEX IF EXISTS webauthn_devices_lookup_key;

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

INSERT INTO webauthn_devices (id, created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning)
SELECT id, created_at, last_used_at, rpid, username, description, kid, public_key, attestation_type, transport, aaguid, sign_count, clone_warning
FROM _bkp_UP_V0003_webauthn_devices;

DROP TABLE IF EXISTS _bkp_UP_V0003_webauthn_devices;
