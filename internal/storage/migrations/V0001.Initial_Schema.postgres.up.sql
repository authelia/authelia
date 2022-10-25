CREATE TABLE authentication_logs (
    id SERIAL,
    time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    successful BOOLEAN NOT NULL,
    banned BOOLEAN NOT NULL DEFAULT FALSE,
    username VARCHAR(100) NOT NULL,
    auth_type VARCHAR(8) NOT NULL DEFAULT '1FA',
    remote_ip VARCHAR(39) NULL DEFAULT NULL,
    request_uri TEXT,
    request_method VARCHAR(8) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX authentication_logs_username_idx ON authentication_logs (time, username, auth_type);
CREATE INDEX authentication_logs_remote_ip_idx ON authentication_logs (time, remote_ip, auth_type);

CREATE TABLE identity_verification (
    id SERIAL,
    jti CHAR(36),
    iat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    issued_ip VARCHAR(39) NOT NULL,
    exp TIMESTAMP WITH TIME ZONE NOT NULL,
    username VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    consumed TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE (jti)
);

CREATE TABLE totp_configurations (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    secret BYTEA NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE u2f_devices (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BYTEA NOT NULL,
    public_key BYTEA NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username, description)
);

CREATE TABLE duo_devices (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    device VARCHAR(32) NOT NULL,
    method VARCHAR(16) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE user_preferences (
    id SERIAL,
    username VARCHAR(100) NOT NULL,
    second_factor_method VARCHAR(11) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE migrations (
    id SERIAL,
    applied TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version_before INTEGER NULL DEFAULT NULL,
    version_after INTEGER NOT NULL,
    application_version VARCHAR(128) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE encryption (
  id SERIAL,
  name VARCHAR(100),
  value BYTEA NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);
