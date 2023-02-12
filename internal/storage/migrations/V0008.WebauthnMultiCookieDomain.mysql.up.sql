DROP INDEX webauthn_devices_lookup_key ON webauthn_devices;
ALTER TABLE webauthn_devices MODIFY COLUMN rpid VARCHAR(512);
CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (rpid, username, description);
