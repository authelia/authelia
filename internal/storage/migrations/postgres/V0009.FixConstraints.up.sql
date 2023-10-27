ALTER TABLE webauthn_devices
    ALTER COLUMN aaguid DROP NOT NULL;

UPDATE webauthn_devices
SET aaguid = NULL
WHERE aaguid = '' OR aaguid = '00000000-00000000-00000000-00000000';
