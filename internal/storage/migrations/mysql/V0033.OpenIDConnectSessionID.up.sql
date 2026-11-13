CREATE TABLE IF NOT EXISTS oauth2_session_id (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    issuer     CHAR(64)     NOT NULL,
    sector_id  VARCHAR(255) NOT NULL,
    public_id  CHAR(36)     NOT NULL,
    sid        CHAR(36)     NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT oauth2_session_id_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX oauth2_session_id_lookup_key  ON oauth2_session_id (issuer, public_id, sector_id);
CREATE UNIQUE INDEX oauth2_session_id_sid_key     ON oauth2_session_id (issuer, sid);

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

CREATE TABLE IF NOT EXISTS oauth2_logout_session (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    flow_id      CHAR(36)     NOT NULL,
    client_id    VARCHAR(255) NOT NULL,
    subject      VARCHAR(255) NOT NULL,
    sid          VARCHAR(255) NOT NULL,
    redirect_uri TEXT         NOT NULL,
    state        TEXT         NOT NULL,
    created_at   TIMESTAMP    NOT NULL,
    expires_at   TIMESTAMP    NOT NULL,
    CONSTRAINT oauth2_logout_session_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX oauth2_logout_session_flow_id_key   ON oauth2_logout_session (flow_id);
CREATE INDEX oauth2_logout_session_expires_at_idx       ON oauth2_logout_session (expires_at);
