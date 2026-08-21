// Package systemd implements the systemd service manager notification protocol.
package systemd

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// NewNotifier returns a Notifier which sends notifications to the service manager socket described by the
// NOTIFY_SOCKET environment variable. If this system does not support service manager notifications this function
// will return nil which is effectively a no-op notifier.
func NewNotifier() (notifier *Notifier, err error) {
	name := os.Getenv(EnvNotifySocket)

	if name == "" {
		return nil, nil
	}

	if !strings.HasPrefix(name, "/") && !strings.HasPrefix(name, "@") {
		return nil, fmt.Errorf("error initializing systemd notifier: environment variable '%s' has an invalid value '%s': the value must be an absolute path or an abstract socket name beginning with '@'", EnvNotifySocket, name)
	}

	var conn *net.UnixConn

	if conn, err = net.DialUnix(netUnixgram, nil, &net.UnixAddr{Name: name, Net: netUnixgram}); err != nil {
		return nil, fmt.Errorf("error initializing systemd notifier: error occurred dialing the socket '%s': %w", name, err)
	}

	return &Notifier{conn: conn}, nil
}

// WatchdogInterval returns the interval at which the service manager expects a notification as described by the
// WATCHDOG_USEC and WATCHDOG_PID environment variables.
func WatchdogInterval() (interval time.Duration, err error) {
	value := os.Getenv(EnvWatchdogUsec)

	if value == "" {
		return 0, nil
	}

	var usec int64

	if usec, err = strconv.ParseInt(value, 10, 64); err != nil {
		return 0, fmt.Errorf("%s: environment variable '%s' has an invalid value '%s': %w", errFmtWatchdog, EnvWatchdogUsec, value, err)
	}

	if usec < 0 {
		return 0, fmt.Errorf("%s: environment variable '%s' has an invalid value '%s': the value must not be negative", errFmtWatchdog, EnvWatchdogUsec, value)
	}

	if usec > maxWatchdogUsec {
		return 0, fmt.Errorf("%s: environment variable '%s' has an invalid value '%s': the value must not exceed %d", errFmtWatchdog, EnvWatchdogUsec, value, maxWatchdogUsec)
	}

	if value = os.Getenv(EnvWatchdogPID); value != "" {
		var pid int

		if pid, err = strconv.Atoi(value); err != nil {
			return 0, fmt.Errorf("%s: environment variable '%s' has an invalid value '%s': %w", errFmtWatchdog, EnvWatchdogPID, value, err)
		}

		if pid != os.Getpid() {
			return 0, nil
		}
	}

	return time.Duration(usec) * time.Microsecond, nil
}

// Notifier sends service manager notifications over a datagram socket using the systemd notification protocol. A nil
// Notifier is valid and all of its methods perform no operation.
type Notifier struct {
	conn *net.UnixConn
}

// Ready notifies the service manager that startup is complete, optionally including a free-form status description.
func (n *Notifier) Ready(status string) (err error) {
	if status == "" {
		return n.Notify(StateReady)
	}

	return n.Notify(StateReady, fmt.Sprintf(fmtStateStatus, status))
}

// Reloading notifies the service manager that a reload has begun. Notifier.Ready must be sent once it's complete.
func (n *Notifier) Reloading() (err error) {
	if usec, ok := monotonic(); ok {
		return n.Notify(StateReloading, fmt.Sprintf(fmtStateMonotonicUsec, usec))
	}

	return n.Notify(StateReloading)
}

// Stopping notifies the service manager that shutdown has begun, optionally including a free-form status description.
func (n *Notifier) Stopping(status string) (err error) {
	if status == "" {
		return n.Notify(StateStopping)
	}

	return n.Notify(StateStopping, fmt.Sprintf(fmtStateStatus, status))
}

// Watchdog notifies the service manager that the process is still alive.
func (n *Notifier) Watchdog() (err error) {
	return n.Notify(StateWatchdog)
}

// Status notifies the service manager of a free-form status description.
func (n *Notifier) Status(status string) (err error) {
	if status == "" {
		return nil
	}

	return n.Notify(fmt.Sprintf(fmtStateStatus, status))
}

// Notify sends the given newline separated states to the service manager.
func (n *Notifier) Notify(states ...string) (err error) {
	if n == nil || n.conn == nil || len(states) == 0 {
		return nil
	}

	if _, err = n.conn.Write([]byte(strings.Join(states, "\n"))); err != nil {
		return fmt.Errorf("error notifying the systemd service manager: %w", err)
	}

	return nil
}

// Close the underlying socket.
func (n *Notifier) Close() (err error) {
	if n == nil || n.conn == nil {
		return nil
	}

	return n.conn.Close()
}
