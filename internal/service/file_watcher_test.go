package service

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestProvisionUsersFileWatcher(t *testing.T) {
	dir := t.TempDir()

	f, err := os.Create(filepath.Join(dir, "users.yml"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	tx, err := templates.New(templates.Config{})
	require.NoError(t, err)

	address, err := schema.NewAddress("tcp://:9091")
	require.NoError(t, err)

	config := &schema.Configuration{
		Server: schema.Server{
			Address: &schema.AddressTCP{Address: *address},
		},
	}

	provision := ProvisionUsersFileWatcher

	ctx := &testCtx{
		Context:       context.Background(),
		Configuration: config,
		Providers: middlewares.Providers{
			Templates: tx,
		},
		Logger: logrus.NewEntry(logging.Logger()),
	}

	watcher, err := provision(ctx)
	assert.NoError(t, err)
	assert.Nil(t, watcher)

	watcher, err = provision(ctx)
	assert.NoError(t, err)
	assert.Nil(t, watcher)

	config.AuthenticationBackend.File = &schema.AuthenticationBackendFile{
		Path:  filepath.Join(dir, "users.yml"),
		Watch: true,
	}

	watcher, err = provision(ctx)
	assert.EqualError(t, err, "error occurred asserting user provider")
	assert.Nil(t, watcher)

	ctx.Providers.UserProvider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)

	config.AuthenticationBackend.File = &schema.AuthenticationBackendFile{
		Watch: true,
	}

	watcher, err = provision(ctx)
	assert.EqualError(t, err, "error initializing file watcher: path must be specified")
	assert.Nil(t, watcher)

	config.AuthenticationBackend.File = &schema.AuthenticationBackendFile{
		Path:  filepath.Join(dir, "users.yml"),
		Watch: true,
	}

	watcher, err = provision(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)
	assert.NotNil(t, watcher.Log())
	assert.Equal(t, "users", watcher.ServiceName())
	assert.Equal(t, "watcher", watcher.ServiceType())

	watcher.Shutdown()
}

func TestNewFileWatcher(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	f, err := os.Create(filepath.Join(dir, "test.log"))
	require.NoError(t, err)

	service, err := NewFileWatcher("example", filepath.Join(dir, "test.log"), reloader, logrus.NewEntry(logging.Logger()))

	assert.NoError(t, err)

	go func() {
		require.NoError(t, service.Run())
	}()

	// Give the service a moment to start.
	time.Sleep(100 * time.Millisecond)

	_, err = f.Write([]byte("test"))
	require.NoError(t, err)

	require.NoError(t, f.Close())

	time.Sleep(time.Second)

	assert.Equal(t, 1, reloader.count)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "test2.log"), []byte("test"), 0600))

	assert.Equal(t, 1, reloader.count)

	assert.NoError(t, os.Remove(filepath.Join(dir, "test2.log")))

	assert.Equal(t, 1, reloader.count)

	service.Shutdown()
}

func TestNewFileWatcherDirectory(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	service, err := NewFileWatcher("example", dir, reloader, logrus.NewEntry(logging.Logger()))

	assert.NoError(t, err)

	go func() {
		require.NoError(t, service.Run())
	}()

	// Give the service a moment to start.
	time.Sleep(100 * time.Millisecond)

	f, err := os.Create(filepath.Join(dir, "test.log"))
	require.NoError(t, err)

	_, err = f.Write([]byte("test"))
	require.NoError(t, err)

	require.NoError(t, f.Close())

	time.Sleep(time.Second)

	assert.Equal(t, 2, reloader.count)

	service.Shutdown()
}

func TestNewFileWatcherBadPath(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	service, err := NewFileWatcher("example", filepath.Join(dir, "test.log"), reloader, logrus.NewEntry(logging.Logger()))

	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(`^error initializing file watcher: error stating file '/tmp/[^/]+/\d+/test.log': file does not exist$`), err.Error())

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

	service, err := NewFileWatcher("example", filepath.Join(dir, "tmp", "test.log"), reloader, logrus.NewEntry(logging.Logger()))

	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(`^error initializing file watcher: error stating file '/tmp/[^/]+/\d+/tmp/test.log': permission denied trying to read the file$`), err.Error())

	require.NoError(t, os.Chmod(filepath.Join(dir, "tmp"), 0o700))

	assert.Nil(t, service)
}

type testReloader struct {
	count  int
	reload bool
	err    error
}

func (r *testReloader) Reload() (bool, error, string) {
	r.count++

	return r.reload, r.err, nil
}
