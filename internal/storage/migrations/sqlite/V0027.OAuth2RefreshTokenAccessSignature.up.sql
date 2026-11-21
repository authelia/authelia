ALTER TABLE oauth2_refresh_token_session ADD COLUMN access_signature VARCHAR(768) NOT NULL DEFAULT '';

UPDATE oauth2_refresh_token_session
SET access_signature = COALESCE((
        SELECT a.signature
        FROM oauth2_access_token_session a
        WHERE a.request_id = oauth2_refresh_token_session.request_id
          AND a.revoked = FALSE), '')
WHERE revoked = FALSE
  AND access_signature = ''
  AND (SELECT COUNT(*)
       FROM oauth2_access_token_session a
       WHERE a.request_id = oauth2_refresh_token_session.request_id
         AND a.revoked = FALSE) = 1;
