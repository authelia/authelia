package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewSQLiteProvider(t *testing.T) {
	dir := t.TempDir()
	testCases := []struct {
		name string
		have *schema.Configuration
	}{
		{
			"ShouldHandleBasic",
			&schema.Configuration{
				Storage: schema.Storage{
					EncryptionKey: "testing-key-only",
					Local: &schema.StorageLocal{
						Path: filepath.Join(dir, "sqlite1.db"),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := NewSQLiteProvider(tc.have)

			assert.NoError(t, err)
			assert.NotNil(t, provider)
		})
	}
}

func TestSQLiteProviderUpsertSession(t *testing.T) {
	provider, err := NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "testing-key-only",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "sqlite.db"),
			},
		},
	})
	require.NoError(t, err)

	assert.NotContains(t, provider.sqlUpsertSession, "REPLACE INTO")
	assert.Contains(t, provider.sqlUpsertSession, "ON CONFLICT (issuer, signature)")

	db, err := sql.Open("sqlite3e", filepath.Join(t.TempDir(), "upsert.db"))
	require.NoError(t, err)

	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE session (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			issuer      CHAR(64) NOT NULL,
			signature   CHAR(64) NOT NULL,
			public_id   CHAR(36) NOT NULL,
			username    VARCHAR(100) NOT NULL,
			expiration  TIMESTAMP NOT NULL,
			data        BLOB NOT NULL
		);
		CREATE UNIQUE INDEX session_signature_key ON session (issuer, signature);
		CREATE UNIQUE INDEX session_public_id_key ON session (issuer, public_id);`)
	require.NoError(t, err)

	expiration := time.Now().Add(time.Hour)

	_, err = db.Exec(provider.sqlUpsertSession, "issuer", "signature", "public", "john", expiration, []byte("first"))
	require.NoError(t, err)

	var id int

	require.NoError(t, db.QueryRow(`SELECT id FROM session WHERE signature = ?;`, "signature").Scan(&id))

	_, err = db.Exec(provider.sqlUpsertSession, "issuer", "signature", "public", "john", expiration, []byte("second"))
	require.NoError(t, err)

	var (
		updated int
		data    []byte
	)

	require.NoError(t, db.QueryRow(`SELECT id, data FROM session WHERE signature = ?;`, "signature").Scan(&updated, &data))

	assert.Equal(t, id, updated)
	assert.Equal(t, []byte("second"), data)
}

func TestSQLiteRegisteredFuncs(t *testing.T) {
	output := sqlite3BLOBToTEXTBase64([]byte("example"))
	assert.Equal(t, "ZXhhbXBsZQ==", output)

	decoded, err := sqlite3TEXTBase64ToBLOB("ZXhhbXBsZQ==")
	assert.NoError(t, err)
	assert.Equal(t, []byte("example"), decoded)
}
