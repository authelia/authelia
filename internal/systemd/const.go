package systemd

import (
	"math"
	"time"
)

// Environment variables used by the systemd service manager notification protocol.
const (
	// EnvNotifySocket is the environment variable containing the path of the service manager notification socket.
	EnvNotifySocket = "NOTIFY_SOCKET"

	// EnvWatchdogUsec is the environment variable containing the watchdog interval in microseconds.
	EnvWatchdogUsec = "WATCHDOG_USEC"

	// EnvWatchdogPID is the environment variable containing the PID the watchdog interval was intended for.
	EnvWatchdogPID = "WATCHDOG_PID"
)

// States which can be sent to the systemd service manager.
const (
	// StateReady indicates startup is complete.
	StateReady = "READY=1"

	// StateReloading indicates a reload has begun and must be followed by StateReady.
	StateReloading = "RELOADING=1"

	// StateStopping indicates shutdown has begun.
	StateStopping = "STOPPING=1"

	// StateWatchdog indicates the process is still alive.
	StateWatchdog = "WATCHDOG=1"
)

const (
	fmtStateStatus        = "STATUS=%s"
	fmtStateMonotonicUsec = "MONOTONIC_USEC=%d"
)

const (
	netUnixgram = "unixgram"

	errFmtWatchdog = "error determining the systemd watchdog interval"

	maxWatchdogUsec = int64(math.MaxInt64) / int64(time.Microsecond)
)
