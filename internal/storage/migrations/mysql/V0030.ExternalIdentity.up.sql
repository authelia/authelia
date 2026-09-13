CREATE TABLE IF NOT EXISTS user_external_identity_links (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    type VARCHAR(32) NOT NULL,
    provider VARCHAR(32) NOT NULL,
    issuer VARCHAR(512) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    remote_username VARCHAR(255) NULL DEFAULT NULL,
    email VARCHAR(255) NULL DEFAULT NULL,
    signature VARCHAR(128) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX user_external_identity_links_subject_key ON user_external_identity_links (type, issuer(191), subject);
CREATE UNIQUE INDEX user_external_identity_links_provider_key ON user_external_identity_links (username, provider);
CREATE INDEX user_external_identity_links_username_idx ON user_external_identity_links (username);
