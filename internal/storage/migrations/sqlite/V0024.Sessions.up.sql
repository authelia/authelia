CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    session_id VARCHAR(128) NOT NULL,
    data BLOB NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

CREATE UNIQUE INDEX sessions_session_id_key ON sessions (session_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
