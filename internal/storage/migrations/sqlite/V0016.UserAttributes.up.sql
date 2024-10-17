CREATE TABLE IF NOT EXISTS user_attributes (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL,
    disabled BOOLEAN NOT NULL DEFAULT 0,
    last_logged_in TIMESTAMP DEFAULT NULL,
    password_change_required BOOLEAN NOT NULL DEFAULT 0,
    last_password_change TIMESTAMP DEFAULT NULL,
    logout_required BOOLEAN NOT NULL DEFAULT 0,
    user_created_at TIMESTAMP DEFAULT NULL,
    UNIQUE (username)
);