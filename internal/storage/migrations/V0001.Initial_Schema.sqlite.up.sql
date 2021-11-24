CREATE TABLE IF NOT EXISTS authentication_logs (
    id INTEGER,
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    successful BOOLEAN NOT NULL,
    username VARCHAR(100) NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX authentication_logs_username_idx ON authentication_logs (time, username);

CREATE TABLE IF NOT EXISTS identity_verification_tokens (
    id INTEGER,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    token VARCHAR(512),
    PRIMARY KEY (id),
    UNIQUE (token)
);

CREATE TABLE IF NOT EXISTS totp_configurations (
    id INTEGER,
    username VARCHAR(100) NOT NULL,
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER(1) NOT NULL DEFAULT 6,
    totp_period INTEGER NOT NULL DEFAULT 30,
    secret BLOB NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS u2f_devices (
    id INTEGER,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BLOB NOT NULL,
    public_key BLOB NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username, description)
);

CREATE TABLE IF NOT EXISTS user_preferences (
    id INTEGER,
    username VARCHAR(100) UNIQUE NOT NULL,
    second_factor_method VARCHAR(11) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER,
    applied TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version_before INTEGER NULL DEFAULT NULL,
    version_after INTEGER NOT NULL,
    application_version VARCHAR(128) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS encryption (
  id INTEGER,
  name VARCHAR(100),
  value BLOB NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);
