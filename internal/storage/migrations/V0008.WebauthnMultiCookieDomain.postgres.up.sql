DROP INDEX webauthn_devices_lookup_key;
ALTER TABLE webauthn_devices ALTER COLUMN rpid SET DATA TYPE VARCHAR(512);
CREATE UNIQUE INDEX webauthn_devices_lookup_key ON webauthn_devices (rpid, username, description);
