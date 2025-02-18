PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS `users` (
    `id` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    `username` VARCHAR(100) NOT NULL UNIQUE,
    `email` VARCHAR(255) NOT NULL UNIQUE,
    `display_name` VARCHAR(100) DEFAULT '',
    `password` BLOB NOT NULL,
    `disabled` INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS `users_groups` (
    `id` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    `username` VARCHAR(100) NOT NULL,
    `groupname` VARCHAR(100) NOT NULL,
    UNIQUE(username, groupname),
    FOREIGN KEY (`username`) REFERENCES `users`(`username`) ON DELETE CASCADE
);
