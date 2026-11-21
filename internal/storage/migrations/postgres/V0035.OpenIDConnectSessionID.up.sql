CREATE TABLE IF NOT EXISTS oauth2_session_id (
    id         BIGSERIAL CONSTRAINT oauth2_session_id_pkey PRIMARY KEY,
    issuer     CHAR(64)     NOT NULL,
    sector_id  VARCHAR(255) NOT NULL,
    public_id  CHAR(36)     NOT NULL,
    sid        CHAR(36)     NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX oauth2_session_id_lookup_key  ON oauth2_session_id (issuer, public_id, sector_id);
CREATE UNIQUE INDEX oauth2_session_id_sid_key     ON oauth2_session_id (sid, issuer);

ALTER TABLE oauth2_access_token_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_authorization_code_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_device_code_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_openid_connect_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_par_context ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_pkce_request_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_refresh_token_session ADD COLUMN session_id CHAR(36) NULL;
ALTER TABLE oauth2_refresh_token_session ADD COLUMN refresh_token_mode VARCHAR(10) NOT NULL DEFAULT '';

CREATE INDEX oauth2_access_token_session_session_id_idx ON oauth2_access_token_session (session_id);
CREATE INDEX oauth2_authorization_code_session_session_id_idx ON oauth2_authorization_code_session (session_id);
CREATE INDEX oauth2_device_code_session_session_id_idx ON oauth2_device_code_session (session_id);
CREATE INDEX oauth2_openid_connect_session_session_id_idx ON oauth2_openid_connect_session (session_id);
CREATE INDEX oauth2_par_context_session_id_idx ON oauth2_par_context (session_id);
CREATE INDEX oauth2_pkce_request_session_session_id_idx ON oauth2_pkce_request_session (session_id);
CREATE INDEX oauth2_refresh_token_session_session_id_idx ON oauth2_refresh_token_session (session_id);

CREATE TABLE IF NOT EXISTS oauth2_session_logout (
    id           BIGSERIAL CONSTRAINT oauth2_session_logout_pkey PRIMARY KEY,
    flow_id      CHAR(36)     NOT NULL,
    client_id    VARCHAR(255) NOT NULL,
    subject      VARCHAR(255) NOT NULL,
    sid          VARCHAR(255) NOT NULL,
    redirect_uri TEXT         NOT NULL,
    state        TEXT         NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at   TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX oauth2_session_logout_flow_id_key   ON oauth2_session_logout (flow_id);
CREATE INDEX oauth2_session_logout_expires_at_idx       ON oauth2_session_logout (expires_at);

CREATE TABLE IF NOT EXISTS oauth2_session_id_client (
    id         BIGSERIAL CONSTRAINT oauth2_session_id_client_pkey PRIMARY KEY,
    issuer     CHAR(64)     NOT NULL,
    public_id  CHAR(36)     NOT NULL,
    sid        CHAR(36)     NOT NULL,
    client_id  VARCHAR(255) NOT NULL,
    subject    VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX oauth2_session_id_client_key       ON oauth2_session_id_client (issuer, sid, client_id);
CREATE INDEX oauth2_session_id_client_public_id_idx    ON oauth2_session_id_client (issuer, public_id);

CREATE TABLE IF NOT EXISTS oauth2_session_logout_task (
    id              BIGSERIAL CONSTRAINT oauth2_session_logout_task_pkey PRIMARY KEY,
    kind            VARCHAR(16)  NOT NULL,
    sid             CHAR(36)     NOT NULL,
    client_id       VARCHAR(255) NOT NULL DEFAULT '',
    subject         VARCHAR(255) NOT NULL DEFAULT '',
    issuer_url      TEXT         NOT NULL,
    attempts        INTEGER      NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX oauth2_session_logout_task_key              ON oauth2_session_logout_task (sid, kind, client_id);
CREATE INDEX oauth2_session_logout_task_next_attempt_at_idx     ON oauth2_session_logout_task (next_attempt_at);
