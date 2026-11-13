CREATE TABLE IF NOT EXISTS oauth2_blacklisted_iss_jti (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    issuer VARCHAR(512) COLLATE utf8mb4_bin NOT NULL,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

CREATE UNIQUE INDEX oauth2_blacklisted_iss_jti_issuer_signature_key ON oauth2_blacklisted_iss_jti (issuer, signature);
CREATE INDEX oauth2_blacklisted_iss_jti_expires_at_idx ON oauth2_blacklisted_iss_jti (expires_at);
