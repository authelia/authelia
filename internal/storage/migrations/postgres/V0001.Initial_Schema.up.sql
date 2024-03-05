CREATE TABLE IF NOT EXISTS authentication_logs (
    id SERIAL CONSTRAINT authentication_logs_pkey PRIMARY KEY,
    time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    successful BOOLEAN NOT NULL,
    banned BOOLEAN NOT NULL DEFAULT FALSE,
    username VARCHAR(100) NOT NULL,
    auth_type VARCHAR(8) NOT NULL DEFAULT '1FA',
    remote_ip VARCHAR(39) NULL DEFAULT NULL,
    request_uri TEXT,
    request_method VARCHAR(8) NOT NULL DEFAULT ''
);

CREATE INDEX authentication_logs_username_idx ON authentication_logs (time, username, auth_type);
CREATE INDEX authentication_logs_remote_ip_idx ON authentication_logs (time, remote_ip, auth_type);

CREATE TABLE IF NOT EXISTS identity_verification (
    id SERIAL CONSTRAINT identity_verification_pkey PRIMARY KEY,
    jti CHAR(36),
    iat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    issued_ip VARCHAR(39) NOT NULL,
    exp TIMESTAMP WITH TIME ZONE NOT NULL,
    username VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    consumed TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL
);

CREATE UNIQUE INDEX identity_verification_jti_key ON identity_verification (jti);

CREATE TABLE IF NOT EXISTS totp_configurations (
    id SERIAL CONSTRAINT totp_configurations_pkey PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    issuer VARCHAR(100),
    algorithm VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    secret BYTEA NOT NULL
);

CREATE UNIQUE INDEX totp_configurations_username_key ON totp_configurations (username);

CREATE TABLE IF NOT EXISTS u2f_devices (
    id SERIAL CONSTRAINT u2f_devices_pkey PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    description VARCHAR(30) NOT NULL DEFAULT 'Primary',
    key_handle BYTEA NOT NULL,
    public_key BYTEA NOT NULL
);

CREATE UNIQUE INDEX u2f_devices_lookup_key ON u2f_devices (username, description);

CREATE TABLE IF NOT EXISTS duo_devices (
    id SERIAL CONSTRAINT duo_devices_pkey PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    device VARCHAR(32) NOT NULL,
    method VARCHAR(16) NOT NULL
);

CREATE UNIQUE INDEX duo_devices_username_key ON duo_devices (username);

CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL CONSTRAINT user_preferences_pkey PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    second_factor_method VARCHAR(11) NOT NULL
);

CREATE UNIQUE INDEX user_preferences_username_key ON user_preferences (username);

CREATE TABLE IF NOT EXISTS migrations (
    id SERIAL CONSTRAINT migrations_pkey PRIMARY KEY,
    applied TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version_before INTEGER NULL DEFAULT NULL,
    version_after INTEGER NOT NULL,
    application_version VARCHAR(128) NOT NULL
);

CREATE TABLE IF NOT EXISTS encryption (
    id SERIAL CONSTRAINT encryption_pkey PRIMARY KEY,
    name VARCHAR(100),
    value BYTEA NOT NULL
);

CREATE UNIQUE INDEX encryption_name_key ON encryption (name);
