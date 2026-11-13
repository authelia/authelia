CREATE TABLE IF NOT EXISTS oauth2_blacklisted_iss_jti (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    issuer VARCHAR(512) NOT NULL,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX oauth2_blacklisted_iss_jti_issuer_signature_key ON oauth2_blacklisted_iss_jti (issuer, signature);
CREATE INDEX oauth2_blacklisted_iss_jti_expires_at_idx ON oauth2_blacklisted_iss_jti (expires_at);
