CREATE TABLE IF NOT EXISTS recovery_codes (
    id SERIAL CONSTRAINT recovery_codes_pkey PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    signature VARCHAR(128) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    consumed_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    revoked_ip VARCHAR(39) NULL DEFAULT NULL
);

CREATE UNIQUE INDEX recovery_codes_signature_idx ON recovery_codes (signature);
CREATE INDEX recovery_codes_username_idx ON recovery_codes (username);
CREATE INDEX recovery_codes_lookup_idx ON recovery_codes (username, consumed_at, revoked_at);
