package systemd

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotifier(t *testing.T) {
	testCases := []struct {
		Name   string
		Socket func(t *testing.T) string
		Nil    bool
		Err    string
	}{
		{
			Name:   "ShouldReturnNilNotifierWhenSocketUnset",
			Socket: nil,
			Nil:    true,
		},
		{
			Name:   "ShouldReturnNilNotifierWhenSocketEmpty",
			Socket: func(t *testing.T) string { return "" },
			Nil:    true,
		},
		{
			Name:   "ShouldErrorOnRelativeSocketPath",
			Socket: func(t *testing.T) string { return "notify.sock" },
			Nil:    true,
			Err:    "error initializing systemd notifier: environment variable 'NOTIFY_SOCKET' has an invalid value 'notify.sock': the value must be an absolute path or an abstract socket name beginning with '@'",
		},
		{
			Name:   "ShouldErrorOnMissingSocket",
			Socket: func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing.sock") },
			Nil:    true,
			Err:    "error initializing systemd notifier: error occurred dialing the socket",
		},
		{
			Name: "ShouldReturnNotifierForSocket",
			Socket: func(t *testing.T) string {
				path, _ := newTestSocket(t)

				return path
			},
			Nil: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Socket != nil {
				t.Setenv(EnvNotifySocket, tc.Socket(t))
			} else {
				require.NoError(t, os.Unsetenv(EnvNotifySocket))
			}

			notifier, err := NewNotifier()

			if tc.Err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.Err)
			}

			if tc.Nil {
				assert.Nil(t, notifier)
			} else {
				require.NotNil(t, notifier)
				assert.NoError(t, notifier.Close())
			}
		})
	}
}

func TestNotifierShouldSendExpectedStates(t *testing.T) {
	testCases := []struct {
		Name     string
		Notify   func(notifier *Notifier) error
		Expected string
		Assert   func(t *testing.T, actual string)
	}{
		{
			Name:     "Ready",
			Notify:   func(notifier *Notifier) error { return notifier.Ready("") },
			Expected: "READY=1",
		},
		{
			Name:     "ReadyWithStatus",
			Notify:   func(notifier *Notifier) error { return notifier.Ready("Authelia is ready") },
			Expected: "READY=1\nSTATUS=Authelia is ready",
		},
		{
			Name:   "Reloading",
			Notify: func(notifier *Notifier) error { return notifier.Reloading() },
			Assert: func(t *testing.T, actual string) {
				if runtime.GOOS != "linux" {
					assert.Equal(t, "RELOADING=1", actual)

					return
				}

				const prefix = "RELOADING=1\nMONOTONIC_USEC="

				require.True(t, strings.HasPrefix(actual, prefix), "expected '%s' to have the prefix '%s'", actual, prefix)

				value, err := strconv.ParseUint(actual[len(prefix):], 10, 64)

				assert.NoError(t, err)
				assert.NotZero(t, value)
			},
		},
		{
			Name:     "Stopping",
			Notify:   func(notifier *Notifier) error { return notifier.Stopping("") },
			Expected: "STOPPING=1",
		},
		{
			Name:     "StoppingWithStatus",
			Notify:   func(notifier *Notifier) error { return notifier.Stopping("Authelia is shutting down") },
			Expected: "STOPPING=1\nSTATUS=Authelia is shutting down",
		},
		{
			Name:     "Watchdog",
			Notify:   func(notifier *Notifier) error { return notifier.Watchdog() },
			Expected: "WATCHDOG=1",
		},
		{
			Name:     "Status",
			Notify:   func(notifier *Notifier) error { return notifier.Status("Shutting down") },
			Expected: "STATUS=Shutting down",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			path, received := newTestSocket(t)

			t.Setenv(EnvNotifySocket, path)

			notifier, err := NewNotifier()

			require.NoError(t, err)
			require.NotNil(t, notifier)

			defer notifier.Close()

			require.NoError(t, tc.Notify(notifier))

			select {
			case actual := <-received:
				if tc.Assert == nil {
					assert.Equal(t, tc.Expected, actual)
				} else {
					tc.Assert(t, actual)
				}
			case <-time.After(time.Second * 5):
				t.Fatal("timeout waiting for the notification")
			}
		})
	}
}

func TestNotifierShouldNotPanicWhenNil(t *testing.T) {
	var notifier *Notifier

	assert.NoError(t, notifier.Ready("Authelia is ready"))
	assert.NoError(t, notifier.Reloading())
	assert.NoError(t, notifier.Stopping("Authelia is shutting down"))
	assert.NoError(t, notifier.Watchdog())
	assert.NoError(t, notifier.Status("example"))
	assert.NoError(t, notifier.Notify("EXAMPLE=1"))
	assert.NoError(t, notifier.Close())
}

func TestNotifierShouldNotSendEmptyStates(t *testing.T) {
	path, received := newTestSocket(t)

	t.Setenv(EnvNotifySocket, path)

	notifier, err := NewNotifier()

	require.NoError(t, err)
	require.NotNil(t, notifier)

	defer notifier.Close()

	require.NoError(t, notifier.Notify())
	require.NoError(t, notifier.Status(""))
	require.NoError(t, notifier.Watchdog())

	select {
	case actual := <-received:
		assert.Equal(t, "WATCHDOG=1", actual)
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for the notification")
	}
}

func TestWatchdogInterval(t *testing.T) {
	testCases := []struct {
		Name     string
		Usec     *string
		PID      *string
		Expected time.Duration
		Err      string
	}{
		{
			Name:     "ShouldBeDisabledWhenUnset",
			Expected: 0,
		},
		{
			Name:     "ShouldBeDisabledWhenEmpty",
			Usec:     testStrPtr(""),
			Expected: 0,
		},
		{
			Name:     "ShouldParseMicroseconds",
			Usec:     testStrPtr("30000000"),
			Expected: time.Second * 30,
		},
		{
			Name:     "ShouldBeDisabledWhenZero",
			Usec:     testStrPtr("0"),
			Expected: 0,
		},
		{
			Name: "ShouldErrorOnInvalidValue",
			Usec: testStrPtr("abc"),
			Err:  "error determining the systemd watchdog interval: environment variable 'WATCHDOG_USEC' has an invalid value 'abc': strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
		{
			Name: "ShouldErrorOnNegativeValue",
			Usec: testStrPtr("-1"),
			Err:  "error determining the systemd watchdog interval: environment variable 'WATCHDOG_USEC' has an invalid value '-1': the value must not be negative",
		},
		{
			Name:     "ShouldParseMaximumValue",
			Usec:     testStrPtr("9223372036854775"),
			Expected: time.Duration(9223372036854775) * time.Microsecond,
		},
		{
			Name: "ShouldErrorOnValueWhichOverflows",
			Usec: testStrPtr("9223372036854776"),
			Err:  "error determining the systemd watchdog interval: environment variable 'WATCHDOG_USEC' has an invalid value '9223372036854776': the value must not exceed 9223372036854775",
		},
		{
			Name:     "ShouldBeEnabledForMatchingPID",
			Usec:     testStrPtr("10000000"),
			PID:      testStrPtr(strconv.Itoa(os.Getpid())),
			Expected: time.Second * 10,
		},
		{
			Name:     "ShouldBeDisabledForMismatchedPID",
			Usec:     testStrPtr("10000000"),
			PID:      testStrPtr(strconv.Itoa(os.Getpid() + 1)),
			Expected: 0,
		},
		{
			Name: "ShouldErrorOnInvalidPID",
			Usec: testStrPtr("10000000"),
			PID:  testStrPtr("abc"),
			Err:  "error determining the systemd watchdog interval: environment variable 'WATCHDOG_PID' has an invalid value 'abc': strconv.Atoi: parsing \"abc\": invalid syntax",
		},
		{
			Name: "ShouldErrorOnOutOfRangePID",
			Usec: testStrPtr("10000000"),
			PID:  testStrPtr("99999999999999999999"),
			Err:  "error determining the systemd watchdog interval: environment variable 'WATCHDOG_PID' has an invalid value '99999999999999999999': strconv.Atoi: parsing \"99999999999999999999\": value out of range",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Usec == nil {
				require.NoError(t, os.Unsetenv(EnvWatchdogUsec))
			} else {
				t.Setenv(EnvWatchdogUsec, *tc.Usec)
			}

			if tc.PID == nil {
				require.NoError(t, os.Unsetenv(EnvWatchdogPID))
			} else {
				t.Setenv(EnvWatchdogPID, *tc.PID)
			}

			actual, err := WatchdogInterval()

			if tc.Err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.Expected, actual)
			} else {
				assert.EqualError(t, err, tc.Err)
				assert.Equal(t, time.Duration(0), actual)
			}
		})
	}
}

func testStrPtr(value string) *string {
	return &value
}

func newTestSocket(t *testing.T) (path string, received <-chan string) {
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
