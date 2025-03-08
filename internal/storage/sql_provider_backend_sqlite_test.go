package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
					Local: &schema.StorageLocal{
						Path: filepath.Join(dir, "sqlite1.db"),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, NewSQLiteProvider(tc.have))
		})
	}
}

func TestSQLiteRegisteredFuncs(t *testing.T) {
	output := sqlite3BLOBToTEXTBase64([]byte("example"))
	assert.Equal(t, "ZXhhbXBsZQ==", output)

	decoded, err := sqlite3TEXTBase64ToBLOB("ZXhhbXBsZQ==")
	assert.NoError(t, err)
	assert.Equal(t, []byte("example"), decoded)
}
