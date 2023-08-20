DELETE FROM oauth2_access_token_session WHERE challenge_id IS NULL OR subject IS NULL;
ALTER TABLE oauth2_access_token_session ALTER COLUMN challenge_id DROP DEFAULT, ALTER COLUMN challenge_id SET NOT NULL, ALTER COLUMN subject DROP DEFAULT, ALTER COLUMN subject SET NOT NULL;
