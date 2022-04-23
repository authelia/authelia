PRAGMA foreign_keys=off;

BEGIN TRANSACTION;

ALTER TABLE oauth2_consent_session RENAME TO _bkp_UP_V0005_oauth2_consent_session;

CREATE TABLE IF NOT EXISTS oauth2_consent_session (
    id INTEGER,
    challenge_id CHAR(36) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    subject CHAR(36) NULL DEFAULT NULL,
    authorized BOOLEAN NOT NULL DEFAULT FALSE,
    granted BOOLEAN NOT NULL DEFAULT FALSE,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP NULL DEFAULT NULL,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    form_data TEXT NOT NULL,
    requested_scopes TEXT NOT NULL,
    granted_scopes TEXT NOT NULL,
    requested_audience TEXT NULL DEFAULT '',
    granted_audience TEXT NULL DEFAULT '',
    PRIMARY KEY (id),
    CONSTRAINT oauth2_consent_session_subject_fkey
        FOREIGN KEY(subject)
            REFERENCES user_opaque_identifier(identifier) ON UPDATE RESTRICT ON DELETE RESTRICT
);


INSERT INTO oauth2_consent_session (id, challenge_id, client_id, subject, authorized, granted, requested_at, responded_at, expires_at, form_data, requested_scopes, granted_scopes, requested_audience, granted_audience)
SELECT id, challenge_id, client_id, subject, authorized, granted, requested_at, responded_at, expires_at, form_data, requested_scopes, granted_scopes, requested_audience, granted_audience
FROM _bkp_UP_V0005_oauth2_consent_session
ORDER BY id;

DROP TABLE _bkp_UP_V0005_oauth2_consent_session;

COMMIT;

PRAGMA foreign_keys=on;
