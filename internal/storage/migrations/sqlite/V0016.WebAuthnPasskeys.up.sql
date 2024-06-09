ALTER TABLE webauthn_credentials
    ADD COLUMN attestation BLOB NULL DEFAULT NULL;
