package mocks

import (
	"testing"

	"github.com/rpadovani/sqlx-v2"
	"github.com/stretchr/testify/require"

	// Register the SQLite driver used to materialize the rows.
	_ "github.com/mattn/go-sqlite3"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

const sqlSchemaOAuth2ConsentPreConfiguration = `
CREATE TABLE oauth2_consent_preconfiguration (
	id INTEGER PRIMARY KEY,
	client_id VARCHAR(255) NOT NULL,
	subject CHAR(36) NOT NULL,
	created_at DATETIME NOT NULL,
	expires_at DATETIME NULL,
	revoked BOOLEAN NOT NULL DEFAULT FALSE,
	scopes VARCHAR(511) NOT NULL,
	audience VARCHAR(511) NULL,
	requested_claims TEXT NULL,
	signature_claims VARCHAR(255) NULL,
	granted_claims TEXT NULL
);`

// NewConsentPreConfigRows materializes the provided consent pre-configurations as a
// *storage.ConsentPreConfigRows backed by an in-memory SQLite database. The rows iterate and scan exactly as the
// storage provider's own rows do, which lets consumers exercise code paths that consume pre-configurations without
// standing up the real storage provider.
//
// The returned closer must be called once the rows have been consumed.
func NewConsentPreConfigRows(t *testing.T, configs ...model.OAuth2ConsentPreConfig) (rows *storage.ConsentPreConfigRows, closer func()) {
	t.Helper()

	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(sqlSchemaOAuth2ConsentPreConfiguration)
	require.NoError(t, err)

	for _, config := range configs {
		_, err = db.NamedExec(`INSERT INTO oauth2_consent_preconfiguration (id, client_id, subject, created_at, expires_at, revoked, scopes, audience, requested_claims, signature_claims, granted_claims)
			VALUES (:id, :client_id, :subject, :created_at, :expires_at, :revoked, :scopes, :audience, :requested_claims, :signature_claims, :granted_claims)`, config)
		require.NoError(t, err)
	}

	r, err := db.Queryx(`SELECT id, client_id, subject, created_at, expires_at, revoked, scopes, audience, requested_claims, signature_claims, granted_claims FROM oauth2_consent_preconfiguration ORDER BY id`)
	require.NoError(t, err)

	return storage.NewConsentPreConfigRows(r), func() {
		_ = r.Close()
		_ = db.Close()
	}
}
