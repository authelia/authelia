// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package systemd

// monotonic returns the current CLOCK_MONOTONIC time in microseconds.
func monotonic() (usec int64, ok bool) {
	return 0, false
}
