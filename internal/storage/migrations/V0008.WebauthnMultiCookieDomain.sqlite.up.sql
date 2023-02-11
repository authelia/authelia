DROP INDEX webauthn_devices_lookup_key;
CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (rpid, username, description);
