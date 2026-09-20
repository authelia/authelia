DROP INDEX IF EXISTS oauth2_access_token_session_session_id_idx;
DROP INDEX IF EXISTS oauth2_authorization_code_session_session_id_idx;
DROP INDEX IF EXISTS oauth2_device_code_session_session_id_idx;
DROP INDEX IF EXISTS oauth2_openid_connect_session_session_id_idx;
DROP INDEX IF EXISTS oauth2_par_context_session_id_idx;
DROP INDEX IF EXISTS oauth2_pkce_request_session_session_id_idx;
DROP INDEX IF EXISTS oauth2_refresh_token_session_session_id_idx;

ALTER TABLE oauth2_access_token_session DROP COLUMN session_id;
ALTER TABLE oauth2_authorization_code_session DROP COLUMN session_id;
ALTER TABLE oauth2_device_code_session DROP COLUMN session_id;
ALTER TABLE oauth2_openid_connect_session DROP COLUMN session_id;
ALTER TABLE oauth2_par_context DROP COLUMN session_id;
ALTER TABLE oauth2_pkce_request_session DROP COLUMN session_id;
ALTER TABLE oauth2_refresh_token_session DROP COLUMN session_id;

DROP TABLE IF EXISTS oauth2_session_id_client;
