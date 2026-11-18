DROP INDEX IF EXISTS oauth2_access_token_session_signature_idx;
DROP INDEX IF EXISTS oauth2_refresh_token_session_signature_idx;
DROP INDEX IF EXISTS oauth2_authorization_code_session_signature_idx;
DROP INDEX IF EXISTS oauth2_openid_connect_session_signature_idx;
DROP INDEX IF EXISTS oauth2_pkce_request_session_signature_idx;

UPDATE oauth2_refresh_token_session SET active = FALSE WHERE grace_deadline IS NOT NULL;

ALTER TABLE oauth2_refresh_token_session DROP COLUMN grace_deadline;
