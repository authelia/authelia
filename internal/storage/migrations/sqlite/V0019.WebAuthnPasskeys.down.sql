ALTER TABLE webauthn_credentials
    DROP COLUMN attestation;

DROP TABLE IF EXISTS cached_data;
