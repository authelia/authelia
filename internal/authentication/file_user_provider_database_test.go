package authentication

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseModel_Read(t *testing.T) {
	model := &FileDatabaseModel{}

	dir := t.TempDir()

	_, err := os.Create(filepath.Join(dir, "users_database.yml"))

	assert.NoError(t, err)

	assert.EqualError(t, model.Read(filepath.Join(dir, "users_database.yml")), "no file content")

	assert.NoError(t, os.Mkdir(filepath.Join(dir, "x"), 0000))

	f := filepath.Join(dir, "x", "users_database.yml")

	assert.EqualError(t, model.Read(f), fmt.Sprintf("failed to read the '%s' file: open %s: permission denied", f, f))

	f = filepath.Join(dir, "schema.yml")

	file, err := os.Create(f)
	assert.NoError(t, err)

	_, err = file.WriteString("users:\n\tjohn: {}")

	assert.NoError(t, err)

	assert.EqualError(t, model.Read(f), "could not parse the YAML database: yaml: line 2: found character that cannot start any token")
}
