CREATE TABLE IF NOT EXISTS oauth2_dpop_proof (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    jti VARCHAR(255) NOT NULL,
    htm VARCHAR(10) NOT NULL,
    htu VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX oauth2_dpop_proof_jti_htm_htu_key ON oauth2_dpop_proof (jti, htm, htu);
CREATE INDEX oauth2_dpop_proof_expires_at_idx ON oauth2_dpop_proof (expires_at);

CREATE TABLE IF NOT EXISTS oauth2_dpop_nonce (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    signature VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX oauth2_dpop_nonce_signature_key ON oauth2_dpop_nonce (signature);
CREATE INDEX oauth2_dpop_nonce_expires_at_idx ON oauth2_dpop_nonce (expires_at);
