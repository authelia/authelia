CREATE TABLE known_ip_addresses (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at DATETIME NULL,
    user_agent VARCHAR(255) NULL,
    location_info VARCHAR(255) NULL,
);

CREATE UNIQUE INDEX known_ip_addresses_lookup_key ON known_ip_addresses (username, ip_address);
CREATE INDEX expired_known_ip_addresses_lookup_key ON  known_ip_addresses (expires_at);

