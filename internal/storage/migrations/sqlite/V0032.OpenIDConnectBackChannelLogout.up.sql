CREATE TABLE IF NOT EXISTS oauth2_session_id_client (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    issuer     CHAR(64)     NOT NULL,
    public_id  CHAR(36)     NOT NULL,
    sid        CHAR(36)     NOT NULL,
    client_id  VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX oauth2_session_id_client_key       ON oauth2_session_id_client (issuer, sid, client_id);
CREATE INDEX oauth2_session_id_client_public_id_idx    ON oauth2_session_id_client (issuer, public_id);

ALTER TABLE oauth2_access_token_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_authorization_code_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_device_code_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_openid_connect_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_par_context ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_pkce_request_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_refresh_token_session ADD COLUMN session_id CHAR(36) NULL;

CREATE INDEX oauth2_access_token_session_session_id_idx ON oauth2_access_token_session (session_id);
CREATE INDEX oauth2_authorization_code_session_session_id_idx ON oauth2_authorization_code_session (session_id);
CREATE INDEX oauth2_device_code_session_session_id_idx ON oauth2_device_code_session (session_id);
CREATE INDEX oauth2_openid_connect_session_session_id_idx ON oauth2_openid_connect_session (session_id);
CREATE INDEX oauth2_par_context_session_id_idx ON oauth2_par_context (session_id);
CREATE INDEX oauth2_pkce_request_session_session_id_idx ON oauth2_pkce_request_session (session_id);
CREATE INDEX oauth2_refresh_token_session_session_id_idx ON oauth2_refresh_token_session (session_id);
