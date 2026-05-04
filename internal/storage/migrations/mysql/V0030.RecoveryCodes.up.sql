CREATE TABLE IF NOT EXISTS recovery_codes (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    signature VARCHAR(128) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    consumed_at TIMESTAMP NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    revoked_at TIMESTAMP NULL DEFAULT NULL,
    revoked_ip VARCHAR(39) NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX recovery_codes_signature_idx ON recovery_codes (signature);
CREATE INDEX recovery_codes_username_idx ON recovery_codes (username);
CREATE INDEX recovery_codes_lookup_idx ON recovery_codes (username, consumed_at, revoked_at);
