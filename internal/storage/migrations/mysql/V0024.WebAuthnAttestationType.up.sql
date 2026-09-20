ALTER TABLE webauthn_credentials RENAME COLUMN attestation_type TO attestation_format;
ALTER TABLE webauthn_credentials ADD COLUMN attestation_type VARCHAR(32) NOT NULL DEFAULT '';
