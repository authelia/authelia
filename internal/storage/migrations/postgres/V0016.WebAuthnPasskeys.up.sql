ALTER TABLE webauthn_credentials
    ADD COLUMN attestation BYTEA NULL DEFAULT NULL;

CREATE TABLE IF NOT EXISTS cached_data (
    id SERIAL CONSTRAINT cached_data_pkey PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    name VARCHAR(20) NOT NULL,
    encrypted BOOLEAN NOT NULL DEFAULT FALSE,
    value BYTEA NOT NULL
);

CREATE UNIQUE INDEX cached_data_name_key ON cached_data (name);
