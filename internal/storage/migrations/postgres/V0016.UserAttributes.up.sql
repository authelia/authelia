CREATE TABLE IF NOT EXISTS user_attributes (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    last_logged_in TIMESTAMP DEFAULT NULL,
    password_change_required BOOLEAN NOT NULL DEFAULT FALSE,
    last_password_change TIMESTAMP DEFAULT NULL,
    logout_required BOOLEAN NOT NULL DEFAULT FALSE,
    user_created_at TIMESTAMP DEFAULT NULL
);