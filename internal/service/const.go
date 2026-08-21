package service

const (
	fmtLogServerListening = "Listening for %s connections on '%s' path '%s'"
)

const (
	logFieldService  = "service"
	logFieldFile     = "file"
	logFieldOP       = "op"
	logFieldInterval = "interval"

	serviceTypeServer   = "server"
	serviceTypeWatcher  = "watcher"
	serviceTypeSignal   = "signal"
	serviceTypeWatchdog = "watchdog"

	serviceNameSystemd = "systemd"
	serviceNameReload  = "reload"
)

const (
	statusReady    = "Authelia is ready"
	statusStopping = "Authelia is shutting down"
)
