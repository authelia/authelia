ALTER TABLE oauth2_consent_session
    ADD COLUMN granted_claims TEXT NULL;

ALTER TABLE oauth2_consent_preconfiguration ADD COLUMN requested_claims TEXT NULL;
ALTER TABLE oauth2_consent_preconfiguration ADD COLUMN signature_claims CHAR(64) NULL;
ALTER TABLE oauth2_consent_preconfiguration ADD COLUMN granted_claims TEXT DEFAULT '';
