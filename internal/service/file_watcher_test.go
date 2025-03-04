package service

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/logging"
)

func TestNewFileWatcher(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	f, err := os.Create(filepath.Join(dir, "test.log"))

	require.NoError(t, err)

	service, err := NewFileWatcher("example", filepath.Join(dir, "test.log"), reloader, logging.Logger().WithFields(logrus.Fields{}))

	assert.NoError(t, err)

	go func() {
		require.NoError(t, service.Run())
	}()

	_, err = f.Write([]byte("test"))
	require.NoError(t, err)

	require.NoError(t, f.Close())

	time.Sleep(time.Second)

	assert.Equal(t, 1, reloader.count)

	service.Shutdown()
}

func TestNewFileWatcherBadPath(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	service, err := NewFileWatcher("example", filepath.Join(dir, "test.log"), reloader, logging.Logger().WithFields(logrus.Fields{}))

	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(`^error stating file '/tmp/[^/]+/\d+/test.log': file does not exist$`), err.Error())

	assert.Nil(t, service)
}

func TestNewFileWatcherBadPermission(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	require.NoError(t, os.Mkdir(filepath.Join(dir, "tmp"), 0700))

	f, err := os.Create(filepath.Join(dir, "tmp", "test.log"))

	require.NoError(t, err)
	require.NoError(t, f.Close())

	require.NoError(t, os.Chmod(filepath.Join(dir, "tmp"), 0o000))

	service, err := NewFileWatcher("example", filepath.Join(dir, "tmp", "test.log"), reloader, logging.Logger().WithFields(logrus.Fields{}))

	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(`^error stating file '/tmp/[^/]+/\d+/tmp/test.log': permission denied trying to read the file$`), err.Error())

	require.NoError(t, os.Chmod(filepath.Join(dir, "tmp"), 0o700))

	assert.Nil(t, service)
}

type testReloader struct {
	count  int
	reload bool
	err    error
}

func (r *testReloader) Reload() (bool, error) {
	r.count++

	return r.reload, r.err
}
