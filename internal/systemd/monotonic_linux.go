//go:build linux

package systemd

import (
	"syscall"
	"time"
	"unsafe"
)

// clockMonotonic is the CLOCK_MONOTONIC clock identifier. Lives here since this is the only file that uses it and it
// is build gated.
const clockMonotonic = 1

// monotonic returns the current CLOCK_MONOTONIC time in microseconds.
func monotonic() (usec int64, ok bool) {
	var ts syscall.Timespec

	if _, _, errno := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, clockMonotonic, uintptr(unsafe.Pointer(&ts)), 0); errno != 0 {
		return 0, false
	}

	return int64(time.Duration(ts.Nano()) / time.Microsecond), true
}
