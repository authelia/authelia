CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL CONSTRAINT sessions_pkey PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX sessions_session_id_key ON sessions (session_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
