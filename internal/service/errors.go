package service

type errWatcher interface {
	error

	WatcherReloadErrorCritical() bool
}
