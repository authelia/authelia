CREATE TABLE IF NOT EXISTS authentication_logs (
    id INTEGER AUTO_INCREMENT,
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    successful BOOL NOT NULL,
    username VARCHAR(100) NOT NULL,
    auth_type VARCHAR(3) NOT NULL DEFAULT '1FA',
    remote_ip VARCHAR(47) NULL DEFAULT NULL,
    request_uri TEXT NOT NULL,
    request_method VARCHAR(4) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX authentication_logs_username_idx ON authentication_logs (time, username, auth_type);
CREATE INDEX authentication_logs_remote_ip_idx ON authentication_logs (time, remote_ip, auth_type);

CREATE TABLE IF NOT EXISTS identity_verification_tokens (
    id INTEGER AUTO_INCREMENT,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    token VARCHAR(512),
    PRIMARY KEY (id),
    UNIQUE KEY (token)
);

CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER AUTO_INCREMENT,
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    prior INTEGER NULL DEFAULT NULL,
    current INTEGER NOT NULL,
    version VARCHAR(128) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS totp_configurations (
    id INTEGER AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    totp_period INTEGER NOT NULL DEFAULT 30,
    secret VARCHAR(64) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (username)
);

CREATE TABLE IF NOT EXISTS u2f_devices (
    id INTEGER AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BLOB NOT NULL,
    public_key BLOB NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (username, description)
);

CREATE TABLE IF NOT EXISTS user_preferences (
    id INTEGER AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    second_factor_method VARCHAR(11) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (username)
);
