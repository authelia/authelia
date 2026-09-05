ALTER TABLE oauth2_access_token_session
    DROP COLUMN requested_resource,
    DROP COLUMN granted_resource;

ALTER TABLE oauth2_authorization_code_session
    DROP COLUMN requested_resource,
    DROP COLUMN granted_resource;

ALTER TABLE oauth2_openid_connect_session
    DROP COLUMN requested_resource,
    DROP COLUMN granted_resource;

ALTER TABLE oauth2_pkce_request_session
    DROP COLUMN requested_resource,
    DROP COLUMN granted_resource;

ALTER TABLE oauth2_refresh_token_session
    DROP COLUMN requested_resource,
    DROP COLUMN granted_resource;

ALTER TABLE oauth2_device_code_session
    DROP COLUMN requested_resource,
    DROP COLUMN granted_resource;

ALTER TABLE oauth2_consent_session
    DROP COLUMN requested_resource,
    DROP COLUMN granted_resource;

ALTER TABLE oauth2_par_context DROP COLUMN resource;

ALTER TABLE oauth2_consent_preconfiguration DROP COLUMN resource;
