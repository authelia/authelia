ALTER TABLE identity_verification
    ADD COLUMN revoked TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL,
    ADD COLUMN revoked_ip VARCHAR(39) NULL DEFAULT NULL;
