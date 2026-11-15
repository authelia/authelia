CREATE TABLE IF NOT EXISTS user_metadata (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    last_logged_in TIMESTAMP DEFAULT NULL,
    last_password_change TIMESTAMP DEFAULT NULL,
    user_created_at TIMESTAMP DEFAULT NULL
);
