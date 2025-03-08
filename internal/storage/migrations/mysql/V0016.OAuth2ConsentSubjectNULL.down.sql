DELETE FROM oauth2_consent_session
WHERE subject IS NULL;

ALTER TABLE oauth2_consent_session
    MODIFY subject CHAR(36) NOT NULL;
