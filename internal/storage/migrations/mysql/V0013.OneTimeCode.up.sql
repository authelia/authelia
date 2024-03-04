CREATE TABLE IF NOT EXISTS one_time_code (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    signature VARCHAR(128) NOT NULL,
    issued TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    issued_ip VARCHAR(39) NOT NULL,
    expires TIMESTAMP NOT NULL,
    username VARCHAR(100) NOT NULL,
    intent VARCHAR(100) NOT NULL,
    consumed TIMESTAMP NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    revoked TIMESTAMP NULL DEFAULT NULL,
    revoked_ip VARCHAR(39) NULL DEFAULT NULL,
    code BLOB NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX one_time_code_signature ON one_time_code (signature, username);
