ALTER TABLE oauth2_device_code_session
    DROP INDEX oauth2_device_code_session_signature_key,
    DROP INDEX oauth2_device_code_session_user_code_signature_key;
