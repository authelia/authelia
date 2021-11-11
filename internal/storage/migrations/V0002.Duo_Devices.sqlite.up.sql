CREATE TABLE IF NOT EXISTS duo_devices (
    id INTEGER,
    username VARCHAR(100) NOT NULL,
    device VARCHAR(32) NOT NULL,
    method VARCHAR(16) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (username, device)
);
