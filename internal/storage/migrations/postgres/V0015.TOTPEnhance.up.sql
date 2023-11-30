CREATE TABLE IF NOT EXISTS totp_history (
    id SERIAL CONSTRAINT one_time_code_pkey PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username VARCHAR(100) NOT NULL,
	step CHAR(128) NOT NULL
);

CREATE UNIQUE INDEX totp_history_lookup_key ON totp_history (username, step);
