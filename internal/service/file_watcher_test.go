package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
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

	providers := middlewares.NewProvidersBasic()

	providers.Templates, err = templates.New(templates.Config{})
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
		Providers:     providers,
		Logger:        logrus.NewEntry(logging.Logger()),
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

	require.NoError(t, err)

	errCh := make(chan error, 1)

	go func() {
		errCh <- service.Run()
	}()

	_, err = f.Write([]byte("test"))
	require.NoError(t, err)

	require.NoError(t, f.Close())

	assert.Eventually(t, func() bool { return reloader.count.Load() >= 1 }, time.Second*5, time.Millisecond*10)

	service.Shutdown()

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("service did not shut down within timeout")
	}
}

func TestNewFileWatcherDirectory(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	service, err := NewFileWatcher("example", dir, reloader, logrus.NewEntry(logging.Logger()))

	require.NoError(t, err)

	errCh := make(chan error, 1)

	go func() {
		errCh <- service.Run()
	}()

	f, err := os.Create(filepath.Join(dir, "test.log"))
	require.NoError(t, err)

	_, err = f.Write([]byte("test"))
	require.NoError(t, err)

	require.NoError(t, f.Close())

	assert.Eventually(t, func() bool { return reloader.count.Load() >= 1 }, time.Second*5, time.Millisecond*10)

	service.Shutdown()

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("service did not shut down within timeout")
	}
}

func TestNewFileWatcherBadPath(t *testing.T) {
	dir := t.TempDir()

	reloader := &testReloader{reload: true}

	service, err := NewFileWatcher("example", filepath.Join(dir, "test.log"), reloader, logrus.NewEntry(logging.Logger()))

	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(`^error initializing file watcher: error stating file '%s/test.log': file does not exist$`, dir)), err.Error())

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
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(`^error initializing file watcher: error stating file '%s/tmp/test.log': permission denied trying to read the file$`, dir)), err.Error())

	require.NoError(t, os.Chmod(filepath.Join(dir, "tmp"), 0o700))

	assert.Nil(t, service)
}

func TestFileWatcherRunShouldHandleReloadOutcomes(t *testing.T) {
	testCases := []struct {
		name     string
		reloader *testReloader
		level    logrus.Level
		expected string
	}{
		{
			name:     "ShouldLogSuccessfulReload",
			reloader: &testReloader{reload: true},
			level:    logrus.InfoLevel,
			expected: "Reloaded successfully",
		},
		{
			name:     "ShouldLogSkippedReload",
			reloader: &testReloader{},
			level:    logrus.DebugLevel,
			expected: "Reload was triggered but it was skipped",
		},
		{
			name:     "ShouldLogReloadError",
			reloader: &testReloader{err: errors.New("failed to reload")},
			level:    logrus.ErrorLevel,
			expected: "Error occurred during reload",
		},
		{
			name:     "ShouldLogCriticalWatcherReloadError",
			reloader: &testReloader{err: &testErrWatcher{critical: true}},
			level:    logrus.ErrorLevel,
			expected: "Error occurred during reload",
		},
		{
			name:     "ShouldLogNonCriticalWatcherReloadError",
			reloader: &testReloader{err: &testErrWatcher{}},
			level:    logrus.DebugLevel,
			expected: "Reload was triggered but it was skipped",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			path := filepath.Join(dir, "test.log")

			require.NoError(t, os.WriteFile(path, []byte("test"), 0600))

			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.TraceLevel)

			service, err := NewFileWatcher("example", path, tc.reloader, logrus.NewEntry(logger))
			require.NoError(t, err)

			errCh := make(chan error, 1)

			go func() {
				errCh <- service.Run()
			}()

			require.NoError(t, os.WriteFile(path, []byte("test test"), 0600))

			assert.Eventually(t, func() bool { return testLogHasEntry(hook, tc.level, tc.expected) }, time.Second*5, time.Millisecond*10)

			service.Shutdown()

			select {
			case err = <-errCh:
				assert.NoError(t, err)
			case <-time.After(time.Second):
				t.Fatal("service did not shut down within timeout")
			}
		})
	}
}

func TestFileWatcherHandleEventShouldIgnoreIrrelevantEvents(t *testing.T) {
	testCases := []struct {
		name     string
		event    fsnotify.Event
		level    logrus.Level
		expected string
	}{
		{
			name:     "ShouldIgnoreEventsForOtherFiles",
			event:    fsnotify.Event{Name: "other.log", Op: fsnotify.Write},
			level:    logrus.TraceLevel,
			expected: "File modification detected to irrelevant file",
		},
		{
			name:     "ShouldIgnoreRemovalOfTheWatchedFile",
			event:    fsnotify.Event{Name: "test.log", Op: fsnotify.Remove},
			level:    logrus.DebugLevel,
			expected: "File remove was detected",
		},
		{
			name:  "ShouldIgnoreChmodOfTheWatchedFile",
			event: fsnotify.Event{Name: "test.log", Op: fsnotify.Chmod},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.TraceLevel)

			reloader := &testReloader{reload: true}

			service := &FileWatcher{
				name:      "example",
				reload:    reloader,
				log:       logrus.NewEntry(logger),
				directory: "/tmp",
				file:      "test.log",
			}

			service.handleEvent(tc.event)

			assert.Equal(t, int32(0), reloader.count.Load())

			if tc.expected != "" {
				assert.True(t, testLogHasEntry(hook, tc.level, tc.expected))
			}
		})
	}
}

func TestFileWatcherHandleErrorShouldLogTheError(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)

	service := &FileWatcher{
		name:      "example",
		reload:    &testReloader{reload: true},
		log:       logrus.NewEntry(logger),
		directory: "/tmp",
		file:      "test.log",
	}

	service.handleError(errors.New("failed to watch"))

	assert.True(t, testLogHasEntry(hook, logrus.ErrorLevel, "Error while watching file for changes"))
}

func TestFileWatcherRunShouldNotReturnStaleReloadErrors(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "test.log")

	require.NoError(t, os.WriteFile(path, []byte("test"), 0600))

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)

	reloader := &testReloader{err: errors.New("failed to reload"), panic: true, panicAfter: 1}

	service, err := NewFileWatcher("example", path, reloader, logrus.NewEntry(logger))
	require.NoError(t, err)

	errCh := make(chan error, 1)

	go func() {
		errCh <- service.Run()
	}()

	require.NoError(t, os.WriteFile(path, []byte("test test"), 0600))

	assert.Eventually(t, func() bool {
		return testLogHasEntry(hook, logrus.ErrorLevel, "Error occurred during reload")
	}, time.Second*5, time.Millisecond*10)

	require.NoError(t, os.WriteFile(path, []byte("test test test"), 0600))

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("service did not return within timeout")
	}

	assert.True(t, testLogHasEntry(hook, logrus.ErrorLevel, "Critical error caught (recovered)"))

	service.Shutdown()
}

func TestFileWatcherRunShouldRecoverFromPanics(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "test.log")

	require.NoError(t, os.WriteFile(path, []byte("test"), 0600))

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)

	service, err := NewFileWatcher("example", path, &testReloader{panic: true}, logrus.NewEntry(logger))
	require.NoError(t, err)

	errCh := make(chan error, 1)

	go func() {
		errCh <- service.Run()
	}()

	require.NoError(t, os.WriteFile(path, []byte("test test"), 0600))

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("service did not return within timeout")
	}

	assert.True(t, testLogHasEntry(hook, logrus.ErrorLevel, "Critical error caught (recovered)"))

	service.Shutdown()
}

func testLogHasEntry(hook *test.Hook, level logrus.Level, message string) bool {
	for _, entry := range hook.AllEntries() {
		if entry.Level == level && entry.Message == message {
			return true
		}
	}

	return false
}

type testErrWatcher struct {
	critical bool
}

func (e *testErrWatcher) Error() string {
	return "failed to reload"
}

func (e *testErrWatcher) WatcherReloadErrorCritical() bool {
	return e.critical
}

func TestNewFileWatcherShouldNotLeakTheWatcherOnFailureToAddPath(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "watched")

	require.NoError(t, os.Mkdir(path, 0o700))
	require.NoError(t, os.Chmod(path, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(path, 0o700)
	})

	reloader := &testReloader{reload: true}
	log := logrus.NewEntry(logging.Logger())

	service, err := NewFileWatcher("example", path, reloader, log)
	if err == nil {
		service.Shutdown()

		t.Skip("unable to force a failure adding the path to the watch list")
	}

	assert.ErrorContains(t, err, fmt.Sprintf("failed to add path '%s' to watch list: ", path))
	assert.Nil(t, service)

	before := testCountOpenFileDescriptors(t)

	for i := 0; i < 20; i++ {
		service, err = NewFileWatcher("example", path, reloader, log)

		require.Error(t, err)
		require.Nil(t, service)
	}

	assert.Less(t, testCountOpenFileDescriptors(t)-before, 20)
}

func testCountOpenFileDescriptors(t *testing.T) int {
	t.Helper()

	path := "/proc/self/fd"

	if runtime.GOOS == "darwin" {
		path = "/dev/fd"
	}

	f, err := os.Open(path)
	if err != nil {
		t.Skipf("unable to determine the number of open file descriptors: %v", err)
	}

	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		t.Skipf("unable to determine the number of open file descriptors: %v", err)
	}

	return len(names)
}

type testReloader struct {
	count      atomic.Int32
	reload     bool
	panic      bool
	panicAfter int32
	err        error
}

func (r *testReloader) Reload() (bool, error) {
	count := r.count.Add(1)

	if r.panic && count > r.panicAfter {
		panic("failed to reload")
	}

	return r.reload, r.err
}
