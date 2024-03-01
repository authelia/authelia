CREATE TABLE IF NOT EXISTS totp_history (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username VARCHAR(100) NOT NULL,
    step CHAR(64) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX totp_history_lookup_key ON totp_history (username, step);
