CREATE TABLE IF NOT EXISTS telegram_verifications (
    id INTEGER NOT NULL AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    token VARCHAR(64) NOT NULL,
    telegram_id BIGINT NOT NULL DEFAULT 0,
    phone VARCHAR(20) NOT NULL DEFAULT '',
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE INDEX telegram_verifications_username_token_idx ON telegram_verifications (username, token);
