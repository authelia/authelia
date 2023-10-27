DELETE FROM oauth2_access_token_session WHERE challenge_id IS NULL OR subject IS NULL;
ALTER TABLE oauth2_access_token_session MODIFY challenge_id CHAR(36) NOT NULL, MODIFY subject CHAR(36) NOT NULL;
