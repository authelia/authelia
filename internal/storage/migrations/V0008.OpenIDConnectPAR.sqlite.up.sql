CREATE TABLE IF NOT EXISTS oauth2_par_context (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    signature VARCHAR(255) NOT NULL,
    request_id VARCHAR(40) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    scopes TEXT NOT NULL,
    audience TEXT NOT NULL,
    handled_response_types TEXT NOT NULL,
    response_mode TEXT NOT NULL,
    response_mode_default TEXT NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    form_data TEXT NOT NULL,
    session_data BLOB NOT NULL
);

CREATE UNIQUE INDEX oauth2_par_context_signature_key ON oauth2_par_context (signature);
