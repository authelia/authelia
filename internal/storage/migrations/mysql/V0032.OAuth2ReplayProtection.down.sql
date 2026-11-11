DROP INDEX oauth2_device_code_session_signature_key ON oauth2_device_code_session;
DROP INDEX oauth2_device_code_session_user_code_signature_key ON oauth2_device_code_session;

DROP TABLE IF EXISTS oauth2_request_replay;

DELETE FROM oauth2_blacklisted_jti
WHERE id NOT IN (
    SELECT id FROM (
        SELECT MAX(id) AS id
        FROM oauth2_blacklisted_jti
        GROUP BY signature
    ) AS latest
);

DROP INDEX oauth2_blacklisted_jti_client_id_signature_key ON oauth2_blacklisted_jti;

CREATE UNIQUE INDEX oauth2_blacklisted_jti_signature_key ON oauth2_blacklisted_jti (signature);

ALTER TABLE oauth2_blacklisted_jti
    DROP COLUMN client_id;
