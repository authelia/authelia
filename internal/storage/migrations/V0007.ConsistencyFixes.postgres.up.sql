DROP TABLE _bkp_UP_V0002_totp_configurations;
DROP TABLE _bkp_UP_V0002_u2f_devices;

ALTER TABLE oauth2_consent_session DROP CONSTRAINT IF EXISTS oauth2_consent_session_subject_fkey;
ALTER TABLE oauth2_consent_session DROP CONSTRAINT IF EXISTS oauth2_consent_session_preconfiguration_fkey;
ALTER TABLE oauth2_consent_preconfiguration DROP CONSTRAINT IF EXISTS oauth2_consent_preconfiguration_subjct_fkey;
ALTER TABLE oauth2_consent_preconfiguration DROP CONSTRAINT IF EXISTS oauth2_consent_preconfiguration_subject_fkey;
ALTER TABLE oauth2_access_token_session DROP CONSTRAINT IF EXISTS oauth2_access_token_session_subject_fkey;
ALTER TABLE oauth2_authorization_code_session DROP CONSTRAINT IF EXISTS oauth2_authorization_code_session_subject_fkey;
ALTER TABLE oauth2_openid_connect_session DROP CONSTRAINT IF EXISTS oauth2_openid_connect_session_subject_fkey;
ALTER TABLE oauth2_pkce_request_session DROP CONSTRAINT IF EXISTS oauth2_pkce_request_session_subject_fkey;
ALTER TABLE oauth2_refresh_token_session DROP CONSTRAINT IF EXISTS oauth2_refresh_token_session_subject_fkey;

ALTER TABLE oauth2_consent_session ADD CONSTRAINT oauth2_consent_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_consent_session ADD CONSTRAINT oauth2_consent_session_preconfiguration_fkey FOREIGN KEY (preconfiguration) REFERENCES oauth2_consent_preconfiguration (id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE oauth2_consent_preconfiguration ADD CONSTRAINT oauth2_consent_preconfiguration_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_access_token_session ADD CONSTRAINT oauth2_access_token_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_authorization_code_session ADD CONSTRAINT oauth2_authorization_code_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_openid_connect_session ADD CONSTRAINT oauth2_openid_connect_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_pkce_request_session ADD CONSTRAINT oauth2_pkce_request_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE oauth2_refresh_token_session ADD CONSTRAINT oauth2_refresh_token_session_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE CASCADE ON DELETE RESTRICT;
