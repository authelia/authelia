ALTER TABLE oauth2_consent_preconfiguration DROP COLUMN dpop_jkt;

DROP TABLE IF EXISTS oauth2_dpop_proof;
DROP TABLE IF EXISTS oauth2_dpop_nonce;
