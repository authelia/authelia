ALTER TABLE oauth2_consent_session ALTER COLUMN subject DROP NOT NULL;
ALTER TABLE oauth2_consent_session ALTER COLUMN subject SET DEFAULT NULL;
ALTER TABLE oauth2_consent_session RENAME CONSTRAINT oauth2_consent_subject_fkey TO oauth2_consent_session_subject_fkey;