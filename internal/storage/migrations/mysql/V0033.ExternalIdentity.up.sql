CREATE TABLE IF NOT EXISTS user_external_identity_links (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    type VARCHAR(32) COLLATE utf8mb4_bin NOT NULL,
    provider VARCHAR(32) COLLATE utf8mb4_bin NOT NULL,
    issuer VARCHAR(512) COLLATE utf8mb4_bin NOT NULL,
    issuer_digest BINARY(32) GENERATED ALWAYS AS (UNHEX(SHA2(issuer, 256))) STORED,
    subject VARCHAR(255) COLLATE utf8mb4_bin NOT NULL,
    username VARCHAR(100) COLLATE utf8mb4_bin NOT NULL,
    remote_username VARCHAR(255) NULL DEFAULT NULL,
    email VARCHAR(255) NULL DEFAULT NULL,
    signature VARCHAR(128) COLLATE utf8mb4_bin NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX user_external_identity_links_subject_key ON user_external_identity_links (type, subject, issuer_digest);
CREATE UNIQUE INDEX user_external_identity_links_provider_key ON user_external_identity_links (username, provider);
CREATE INDEX user_external_identity_links_username_idx ON user_external_identity_links (username);
