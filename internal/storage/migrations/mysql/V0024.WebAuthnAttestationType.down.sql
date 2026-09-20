ALTER TABLE webauthn_credentials DROP COLUMN attestation_type;
ALTER TABLE webauthn_credentials RENAME COLUMN attestation_format TO attestation_type;
