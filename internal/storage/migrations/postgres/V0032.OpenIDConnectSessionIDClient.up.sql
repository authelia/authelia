CREATE TABLE IF NOT EXISTS oauth2_session_id_client (
    id         BIGSERIAL CONSTRAINT oauth2_session_id_client_pkey PRIMARY KEY,
    issuer     CHAR(64)     NOT NULL,
    public_id  CHAR(36)     NOT NULL,
    sid        CHAR(36)     NOT NULL,
    client_id  VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX oauth2_session_id_client_key       ON oauth2_session_id_client (issuer, sid, client_id);
CREATE INDEX oauth2_session_id_client_public_id_idx    ON oauth2_session_id_client (issuer, public_id);
