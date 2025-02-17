CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(100) DEFAULT '',
    password BYTEA NOT NULL,
    disabled BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS users_groups (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    groupname VARCHAR(100) NOT NULL,
    UNIQUE(username, groupname),
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);
