ALTER TABLE oauth2_consent_session
    MODIFY subject CHAR(36) NULL DEFAULT NULL;
