CREATE TABLE IF NOT EXISTS authentication_logs (
    id SERIAL,
    time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    successful BOOLEAN NOT NULL,
    username VARCHAR(100) NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX authentication_logs_username_idx ON authentication_logs (time, username);

CREATE TABLE IF NOT EXISTS identity_verification_tokens (
    id SERIAL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    token VARCHAR(512),
    PRIMARY KEY (id),
    UNIQUE (token)
);

CREATE TABLE IF NOT EXISTS totp_configurations (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    totp_period INTEGER NOT NULL DEFAULT 30,
    secret BYTEA NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS u2f_devices (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BYTEA NOT NULL,
    public_key BYTEA NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username, description)
);

CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    second_factor_method VARCHAR(11) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS migrations (
    id SERIAL,
    applied TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version_before INTEGER NULL DEFAULT NULL,
    version_after INTEGER NOT NULL,
    application_version VARCHAR(128) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS encryption (
  id SERIAL,
  name VARCHAR(100),
  value BYTEA NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);