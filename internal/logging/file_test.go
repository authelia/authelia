package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFile(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, "foo"), 0000)
	require.NoError(t, err)

	l := NewFile(filepath.Join(dir, "authelia.log"))

	assert.EqualError(t, l.Reopen(), "error reopening log file: file isn't open")
	assert.EqualError(t, l.Close(), "error closing log file: file isn't open")
	require.NoError(t, l.Open())
	assert.EqualError(t, l.Open(), "error opening log file: file is already open")
	require.NoError(t, l.Reopen())
	assert.NoError(t, l.Close())
	assert.NoError(t, l.Open())

	d := NewFile(filepath.Join(dir, "foo", "authelia.log"))

	assert.EqualError(t, d.Open(), fmt.Sprintf("error opening log file: open %s: permission denied", filepath.Join(dir, "foo", "authelia.log")))

	l = NewFile(filepath.Join(dir, "authelia2.log"))
	assert.NoError(t, l.Open())

	assert.NoError(t, os.Chmod(filepath.Join(dir, "authelia2.log"), 0000))

	assert.EqualError(t, l.Reopen(), fmt.Sprintf("error reopning log file: error opening new log file: open %s: permission denied", filepath.Join(dir, "authelia2.log")))
}
