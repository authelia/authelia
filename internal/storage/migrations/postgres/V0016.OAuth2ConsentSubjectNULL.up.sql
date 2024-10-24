ALTER TABLE oauth2_consent_session
    ALTER COLUMN subject DROP NOT NULL,
    ALTER COLUMN subject SET DEFAULT NULL;
