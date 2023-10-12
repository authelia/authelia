CREATE TABLE IF NOT EXISTS one_time_code (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	public_id CHAR(36) NOT NULL,
    signature VARCHAR(128) NOT NULL,
	issued DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    issued_ip VARCHAR(39) NOT NULL,
	expires DATETIME NOT NULL,
    username VARCHAR(100) NOT NULL,
	intent VARCHAR(100) NOT NULL,
    consumed DATETIME NULL DEFAULT NULL,
    consumed_ip VARCHAR(39) NULL DEFAULT NULL,
    revoked DATETIME NULL DEFAULT NULL,
    revoked_ip VARCHAR(39) NULL DEFAULT NULL,
	code BLOB NOT NULL
);

CREATE UNIQUE INDEX one_time_code_lookup_key ON one_time_code (signature, username);
CREATE INDEX one_time_code_lookup ON one_time_code (signature, username);

CREATE TABLE IF NOT EXISTS user_elevated_session (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_ip VARCHAR(39) NOT NULL,
	method VARCHAR(10) NOT NULL,
	method_id INTEGER NULL,
    expires DATETIME NOT NULL,
    username VARCHAR(100) NOT NULL
);

CREATE INDEX user_elevated_session_username ON user_elevated_session (username);
