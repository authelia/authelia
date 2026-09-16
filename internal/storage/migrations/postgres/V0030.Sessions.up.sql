CREATE TABLE IF NOT EXISTS session (
    id          SERIAL CONSTRAINT session_pkey PRIMARY KEY,
    issuer      CHAR(64) NOT NULL,
    signature   CHAR(64) NOT NULL,
    public_id   CHAR(36) NOT NULL,
    username    VARCHAR(100) NOT NULL,
    expiration  TIMESTAMP WITH TIME ZONE NOT NULL,
    data        BYTEA NOT NULL
);

CREATE UNIQUE INDEX session_signature_key  ON session (issuer, signature);
CREATE UNIQUE INDEX session_public_id_key  ON session (issuer, public_id);
CREATE INDEX        session_username_idx   ON session (username);
CREATE INDEX        session_expiration_idx ON session (expiration);
