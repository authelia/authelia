ALTER TABLE oauth2_consent_session
    DROP COLUMN granted_claims;

ALTER TABLE oauth2_consent_preconfiguration
    DROP COLUMN requested_claims,
    DROP COLUMN signature_claims,
    DROP COLUMN granted_claims;
