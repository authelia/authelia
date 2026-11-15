DELETE FROM oauth2_device_code_session WHERE challenge_id IS NULL OR subject IS NULL;
ALTER TABLE oauth2_device_code_session
    ALTER COLUMN challenge_id DROP DEFAULT,
    ALTER COLUMN challenge_id SET NOT NULL,
    ALTER COLUMN subject DROP DEFAULT,
    ALTER COLUMN subject SET NOT NULL;
