DELETE FROM oauth2_consent_session
WHERE subject IS NULL;

ALTER TABLE oauth2_consent_session
    ALTER COLUMN subject SET NOT NULL,
    ALTER COLUMN subject DROP DEFAULT;
