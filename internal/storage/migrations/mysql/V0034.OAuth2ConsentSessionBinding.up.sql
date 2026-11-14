ALTER TABLE oauth2_consent_session
    ADD COLUMN binding VARCHAR(64) NOT NULL DEFAULT '' AFTER subject;
