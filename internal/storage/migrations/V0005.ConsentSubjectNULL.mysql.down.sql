ALTER TABLE oauth2_consent_session
DROP FOREIGN KEY oauth2_consent_session_subject_fkey,
    ADD CONSTRAINT oauth2_consent_subject_fkey FOREIGN KEY (subject) REFERENCES user_opaque_identifier (identifier) ON UPDATE RESTRICT ON DELETE RESTRICT;
