DELETE FROM oauth2_consent_session
       WHERE subject IN(SELECT identifier FROM user_opaque_identifier WHERE username = '' AND service IN('openid', 'openid_connect'));

DELETE FROM user_opaque_identifier
       WHERE service <> 'openid' OR username = '';

ALTER TABLE oauth2_consent_session
    ALTER COLUMN subject DROP NOT NULL,
    ALTER COLUMN subject SET DEFAULT NULL;

ALTER TABLE oauth2_consent_session
    RENAME CONSTRAINT oauth2_consent_subject_fkey TO oauth2_consent_session_subject_fkey;
