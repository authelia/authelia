CREATE TABLE IF NOT EXISTS telegram_verifications (
    id SERIAL CONSTRAINT telegram_verifications_pkey PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    token VARCHAR(64) NOT NULL,
    telegram_id BIGINT NOT NULL DEFAULT 0,
    phone VARCHAR(20) NOT NULL DEFAULT '',
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX telegram_verifications_token_unique ON telegram_verifications (token);
CREATE INDEX telegram_verifications_username_idx ON telegram_verifications (username);
