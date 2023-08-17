ALTER TABLE webauthn_devices
	DROP CONSTRAINT IF EXISTS webauthn_devices_pkey;

DROP INDEX IF EXISTS webauthn_devices_pkey;
DROP INDEX IF EXISTS webauthn_devices_kid_key;
DROP INDEX IF EXISTS webauthn_devices_lookup_key;

ALTER TABLE webauthn_devices
    RENAME TO _bkp_UP_V0008_webauthn_devices;

CREATE TABLE IF NOT EXISTS webauthn_devices (
    id SERIAL CONSTRAINT webauthn_devices_pkey PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
	rpid VARCHAR(512) NOT NULL,
    username VARCHAR(100) NOT NULL,
	description VARCHAR(64) NOT NULL,
    kid VARCHAR(512) NOT NULL,
	aaguid CHAR(36) NOT NULL,
    attestation_type VARCHAR(32),
	attachment VARCHAR(64) NOT NULL,
    transport VARCHAR(20) DEFAULT '',
    sign_count INTEGER DEFAULT 0,
    clone_warning BOOLEAN NOT NULL DEFAULT FALSE,
    legacy BOOLEAN NOT NULL DEFAULT FALSE,
	discoverable BOOLEAN NOT NULL,
	present BOOLEAN NOT NULL DEFAULT FALSE,
	verified BOOLEAN NOT NULL DEFAULT FALSE,
	backup_eligible BOOLEAN NOT NULL DEFAULT FALSE,
	backup_state BOOLEAN NOT NULL DEFAULT FALSE,
	public_key BYTEA NOT NULL
);

CREATE UNIQUE INDEX webauthn_devices_kid_key ON webauthn_devices (kid);
CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (rpid, username, description);

INSERT INTO webauthn_devices (created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, attachment, transport, sign_count, clone_warning, legacy, discoverable, present, verified, backup_eligible, backup_state, public_key)
SELECT created_at, last_used_at, rpid, username, description, kid, aaguid, attestation_type, 'cross-platform', transport, sign_count, clone_warning, TRUE, FALSE, FALSE, FALSE, FALSE, FALSE, public_key
FROM _bkp_UP_V0008_webauthn_devices;

DROP TABLE IF EXISTS _bkp_UP_V0008_webauthn_devices;

CREATE TABLE IF NOT EXISTS webauthn_users (
    id SERIAL CONSTRAINT webauthn_users_pkey PRIMARY KEY,
    rpid VARCHAR(512) NOT NULL,
    username VARCHAR(100) NOT NULL,
	userid CHAR(64) NOT NULL
);

CREATE UNIQUE INDEX webauthn_users_lookup_key ON webauthn_users (rpid, username);
