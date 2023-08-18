ALTER TABLE oauth2_access_token_session MODIFY challenge_id CHAR(36) NULL DEFAULT NULL, MODIFY subject CHAR(36) NULL DEFAULT NULL;
