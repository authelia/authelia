CREATE TABLE IF NOT EXISTS totp_history (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username VARCHAR(100) NOT NULL,
    step CHAR(64) NOT NULL
);

CREATE UNIQUE INDEX totp_history_lookup_key ON totp_history (username, step);
