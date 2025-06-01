package service

import "errors"

const (
	fmtLogServerListening = "Listening for %s connections on '%s' path '%s'"
)

const (
	logFieldService = "service"
	logFieldFile    = "file"
	logFieldOP      = "op"

	serviceTypeServer  = "server"
	serviceTypeWatcher = "watcher"
	serviceTypeSignal  = "signal"
)

var (
	// ErrApplicationReload is emitted when the application is being reloaded. This effectively starts Authelia again
	// instead of doing a full exit.
	ErrApplicationReload = errors.New("application reload")
)
