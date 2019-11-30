package utils

import "errors"

// ErrTimeoutReached error thrown when a timeout is reached
var ErrTimeoutReached = errors.New("timeout reached")
