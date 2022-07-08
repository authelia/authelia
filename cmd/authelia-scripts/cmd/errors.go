package cmd

import (
	"errors"
)

// ErrNotAvailableSuite error raised when suite is not available.
var ErrNotAvailableSuite = errors.New("unavailable suite")

// ErrNoRunningSuite error raised when no suite is running.
var ErrNoRunningSuite = errors.New("no running suite")
