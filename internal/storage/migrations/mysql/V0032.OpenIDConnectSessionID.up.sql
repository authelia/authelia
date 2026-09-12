CREATE TABLE IF NOT EXISTS oauth2_session_id (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    issuer     CHAR(64)     NOT NULL,
    sector_id  VARCHAR(255) NOT NULL,
    public_id  CHAR(36)     NOT NULL,
    sid        CHAR(36)     NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT oauth2_session_id_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX oauth2_session_id_lookup_key  ON oauth2_session_id (issuer, public_id, sector_id);
CREATE UNIQUE INDEX oauth2_session_id_sid_key     ON oauth2_session_id (issuer, sid);
