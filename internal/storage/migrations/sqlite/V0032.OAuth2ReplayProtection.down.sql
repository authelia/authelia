DROP INDEX IF EXISTS oauth2_device_code_session_signature_key;
DROP INDEX IF EXISTS oauth2_device_code_session_user_code_signature_key;

DROP TABLE IF EXISTS oauth2_request_replay;

DELETE FROM oauth2_blacklisted_jti
WHERE id NOT IN (
    SELECT MAX(id)
    FROM oauth2_blacklisted_jti
    GROUP BY signature
);

DROP INDEX IF EXISTS oauth2_blacklisted_jti_client_id_signature_key;

CREATE UNIQUE INDEX oauth2_blacklisted_jti_signature_key ON oauth2_blacklisted_jti (signature);

ALTER TABLE oauth2_blacklisted_jti
    DROP COLUMN client_id;
