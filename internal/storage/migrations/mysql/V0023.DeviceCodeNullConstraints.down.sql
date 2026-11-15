DELETE FROM oauth2_device_code_session
WHERE subject IS NULL;

ALTER TABLE oauth2_device_code_session
    MODIFY subject CHAR(36) NOT NULL;
