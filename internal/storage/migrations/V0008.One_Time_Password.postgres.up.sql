CREATE TABLE IF NOT EXISTS one_time_password (
    id SERIAL CONSTRAINT one_time_password_pkey PRIMARY KEY,
	public_id CHAR(36) NOT NULL,
    signature VARCHAR(128) NOT NULL,
    iat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    issued_ip VARCHAR(39) NOT NULL,
    exp TIMESTAMP WITH TIME ZONE NOT NULL,
    username VARCHAR(100) NOT NULL,
	intent VARCHAR(100) NOT NULL,
    consumed TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    revoked TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
	revoked_ip VARCHAR(39) NULL DEFAULT NULL,
    password BYTEA NOT NULL
);

CREATE UNIQUE INDEX one_time_password_lookup_key ON one_time_password (signature, username);
CREATE INDEX one_time_password_lookup ON one_time_password (signature, username);

CREATE TABLE IF NOT EXISTS user_elevated_session (
    id SERIAL CONSTRAINT user_elevated_session_pkey PRIMARY KEY,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_ip VARCHAR(39) NOT NULL,
	method VARCHAR(10) NOT NULL,
	method_id INTEGER NULL,
    expires TIMESTAMP WITH TIME ZONE NOT NULL,
    username VARCHAR(100) NOT NULL
);

CREATE INDEX user_elevated_session_username ON user_elevated_session (username);
