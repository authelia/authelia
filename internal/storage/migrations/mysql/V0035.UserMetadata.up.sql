CREATE TABLE IF NOT EXISTS user_metadata (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    last_logged_in TIMESTAMP DEFAULT NULL,
    last_password_change TIMESTAMP DEFAULT NULL,
    user_created_at TIMESTAMP DEFAULT NULL,
    UNIQUE KEY (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;
