CREATE TABLE authentication_logs (
    id INTEGER,
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
    id INTEGER,
    jti VARCHAR(36),
    iat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    issued_ip VARCHAR(39) NOT NULL,
    exp TIMESTAMP NOT NULL,
    username VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    consumed TIMESTAMP NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE (jti)
);

CREATE TABLE totp_configurations (
    id INTEGER,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    secret BLOB NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE u2f_devices (
    id INTEGER,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BLOB NOT NULL,
    public_key BLOB NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username, description)
);

CREATE TABLE duo_devices (
    id INTEGER,
    username VARCHAR(100) NOT NULL,
    device VARCHAR(32) NOT NULL,
    method VARCHAR(16) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE TABLE user_preferences (
    id INTEGER,
    username VARCHAR(100) UNIQUE NOT NULL,
    second_factor_method VARCHAR(11) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE migrations (
    id INTEGER,
    applied TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version_before INTEGER NULL DEFAULT NULL,
    version_after INTEGER NOT NULL,
    application_version VARCHAR(128) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE encryption (
  id INTEGER,
  name VARCHAR(100),
  value BLOB NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);
