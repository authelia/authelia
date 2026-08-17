package service

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/systemd"
)

func TestProvisionSystemdWatchdog(t *testing.T) {
	testCases := []struct {
		name     string
		usec     string
		socket   bool
		expected bool
		err      string
	}{
		{
			name:     "ShouldNotProvisionWhenWatchdogDisabled",
			usec:     "",
			socket:   true,
			expected: false,
		},
		{
			name:     "ShouldNotProvisionWhenNotifySocketUnset",
			usec:     "30000000",
			socket:   false,
			expected: false,
		},
		{
			name:     "ShouldProvisionWhenEnabled",
			usec:     "30000000",
			socket:   true,
			expected: true,
		},
		{
			name:   "ShouldErrorOnInvalidInterval",
			usec:   "abc",
			socket: true,
			err:    "error determining the systemd watchdog interval: environment variable 'WATCHDOG_USEC' has an invalid value 'abc': strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, os.Unsetenv(systemd.EnvWatchdogPID))

			if tc.usec == "" {
				require.NoError(t, os.Unsetenv(systemd.EnvWatchdogUsec))
			} else {
				t.Setenv(systemd.EnvWatchdogUsec, tc.usec)
			}

			if tc.socket {
				path, _ := newTestNotifySocket(t)

				t.Setenv(systemd.EnvNotifySocket, path)
			} else {
				require.NoError(t, os.Unsetenv(systemd.EnvNotifySocket))
			}

			service, err := ProvisionSystemdWatchdog(newMockServiceCtx())

			if tc.err == "" {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			if tc.expected {
				require.NotNil(t, service)
				assert.Equal(t, serviceTypeWatchdog, service.ServiceType())
				assert.Equal(t, "systemd", service.ServiceName())
				assert.NotNil(t, service.Log())

				service.Shutdown()
			} else {
				assert.Nil(t, service)
			}
		})
	}
}

func TestWatchdogShouldNotifyAtHalfTheInterval(t *testing.T) {
	path, received := newTestNotifySocket(t)

	t.Setenv(systemd.EnvNotifySocket, path)
	t.Setenv(systemd.EnvWatchdogUsec, strconv.Itoa(200000))

	require.NoError(t, os.Unsetenv(systemd.EnvWatchdogPID))

	service, err := ProvisionSystemdWatchdog(newMockServiceCtx())

	require.NoError(t, err)
	require.NotNil(t, service)

	go func() {
		_ = service.Run()
	}()

	for i := 0; i < 2; i++ {
		select {
		case actual := <-received:
			assert.Equal(t, systemd.StateWatchdog, actual)
		case <-time.After(time.Second * 5):
			t.Fatal("timeout waiting for the watchdog notification")
		}
	}

	service.Shutdown()
}

func TestWatchdogShouldShutdownWithoutNotifications(t *testing.T) {
	path, _ := newTestNotifySocket(t)

	t.Setenv(systemd.EnvNotifySocket, path)

	notifier, err := systemd.NewNotifier()

	require.NoError(t, err)

	logger := logrus.New()

	service := &Watchdog{
		interval: time.Hour,
		notifier: notifier,
		log:      logrus.NewEntry(logger),
		quit:     make(chan struct{}),
	}

	done := make(chan struct{})

	go func() {
		assert.NoError(t, service.Run())

		close(done)
	}()

	time.Sleep(time.Millisecond * 100)

	service.Shutdown()

	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Fatal("service did not shut down within timeout")
	}
}

func newTestNotifySocket(t *testing.T) (path string, received <-chan string) {
	dir, err := os.MkdirTemp("", "authelia-systemd")

	require.NoError(t, err)

	path = filepath.Join(dir, "notify.sock")

	if len(path) > 100 {
		t.Skipf("skipping as the socket path '%s' exceeds the maximum length for a unix socket", path)
	}

	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: path, Net: "unixgram"})

	require.NoError(t, err)

	ch := make(chan string, 10)

	t.Cleanup(func() {
		_ = conn.Close()
		_ = os.RemoveAll(dir)
	})

	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if err != nil {
				close(ch)

				return
			}

			ch <- string(buf[:n])
		}
	}()

	return path, ch
}
