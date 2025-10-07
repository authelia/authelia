ALTER TABLE oauth2_device_code_session
    MODIFY subject CHAR(36) NULL DEFAULT NULL;
