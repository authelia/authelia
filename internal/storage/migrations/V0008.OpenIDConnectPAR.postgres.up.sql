CREATE TABLE IF NOT EXISTS oauth2_par_context (
    id SERIAL CONSTRAINT oauth2_par_context_pkey PRIMARY KEY,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    signature VARCHAR(255) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    scopes TEXT NOT NULL,
    audience TEXT NULL DEFAULT '',
    handled_response_types TEXT NOT NULL DEFAULT '',
    response_mode TEXT NOT NULL DEFAULT '',
    response_mode_default TEXT NOT NULL DEFAULT '',
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BYTEA NOT NULL
);

CREATE UNIQUE INDEX oauth2_par_context_signature_key ON oauth2_par_context (signature);
