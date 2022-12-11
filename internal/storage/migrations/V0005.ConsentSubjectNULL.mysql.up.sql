DELETE FROM oauth2_consent_session
       WHERE subject IN(SELECT identifier FROM user_opaque_identifier WHERE username = '' AND service IN('openid', 'openid_connect'));

DELETE FROM user_opaque_identifier
       WHERE username = '' AND service IN('openid', 'openid_connect');

DELETE FROM user_opaque_identifier
       WHERE service <> 'openid';

ALTER TABLE oauth2_consent_session
    MODIFY subject CHAR(36) NULL DEFAULT NULL;

ALTER TABLE oauth2_consent_session
    DROP FOREIGN KEY oauth2_consent_subject_fkey,
    ADD CONSTRAINT oauth2_consent_session_subject_fkey
        FOREIGN KEY (subject)
            REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;
