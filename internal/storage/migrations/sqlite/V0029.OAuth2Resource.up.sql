ALTER TABLE oauth2_access_token_session ADD COLUMN requested_resource TEXT NULL DEFAULT '';
ALTER TABLE oauth2_access_token_session ADD COLUMN granted_resource TEXT NULL DEFAULT '';

ALTER TABLE oauth2_authorization_code_session ADD COLUMN requested_resource TEXT NULL DEFAULT '';
ALTER TABLE oauth2_authorization_code_session ADD COLUMN granted_resource TEXT NULL DEFAULT '';

ALTER TABLE oauth2_openid_connect_session ADD COLUMN requested_resource TEXT NULL DEFAULT '';
ALTER TABLE oauth2_openid_connect_session ADD COLUMN granted_resource TEXT NULL DEFAULT '';

ALTER TABLE oauth2_pkce_request_session ADD COLUMN requested_resource TEXT NULL DEFAULT '';
ALTER TABLE oauth2_pkce_request_session ADD COLUMN granted_resource TEXT NULL DEFAULT '';

ALTER TABLE oauth2_refresh_token_session ADD COLUMN requested_resource TEXT NULL DEFAULT '';
ALTER TABLE oauth2_refresh_token_session ADD COLUMN granted_resource TEXT NULL DEFAULT '';

ALTER TABLE oauth2_device_code_session ADD COLUMN requested_resource TEXT NULL DEFAULT '';
ALTER TABLE oauth2_device_code_session ADD COLUMN granted_resource TEXT NULL DEFAULT '';

ALTER TABLE oauth2_consent_session ADD COLUMN requested_resource TEXT NULL DEFAULT '';
ALTER TABLE oauth2_consent_session ADD COLUMN granted_resource TEXT NULL DEFAULT '';

ALTER TABLE oauth2_par_context ADD COLUMN resource TEXT NOT NULL DEFAULT '';

ALTER TABLE oauth2_consent_preconfiguration ADD COLUMN resource TEXT NULL;
