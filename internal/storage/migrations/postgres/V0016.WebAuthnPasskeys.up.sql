ALTER TABLE webauthn_credentials
    ADD COLUMN attestation BYTEA NULL DEFAULT NULL;
