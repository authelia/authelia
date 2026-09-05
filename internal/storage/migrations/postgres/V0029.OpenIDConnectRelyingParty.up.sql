CREATE TABLE IF NOT EXISTS user_openid_connect_links (
    id SERIAL CONSTRAINT user_openid_connect_links_pkey PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    provider VARCHAR(32) NOT NULL,
    issuer VARCHAR(512) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    remote_username VARCHAR(255) NULL DEFAULT NULL
);

CREATE UNIQUE INDEX user_openid_connect_links_subject_key ON user_openid_connect_links (issuer, subject);
CREATE UNIQUE INDEX user_openid_connect_links_provider_key ON user_openid_connect_links (username, provider);
CREATE INDEX user_openid_connect_links_username_idx ON user_openid_connect_links (username);
