CREATE TABLE known_ip_addresses (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    first_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,

    browser_name VARCHAR(50),
    browser_version VARCHAR(20),
    os_name VARCHAR(50),
    os_version VARCHAR(20),
    device_type VARCHAR(20)
);

CREATE UNIQUE INDEX known_ip_addresses_lookup_key ON known_ip_addresses (username, ip_address);
CREATE INDEX expired_known_ip_addresses_lookup_key ON known_ip_addresses (expires_at);

CREATE OR REPLACE FUNCTION update_last_seen_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_seen = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_known_ip_last_seen
    BEFORE UPDATE ON known_ip_addresses
    FOR EACH ROW
    EXECUTE FUNCTION update_last_seen_column();
