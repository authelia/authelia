package service

import (
	"time"
)

const (
	fmtLogServerListening = "Listening for %s connections on '%s' path '%s'"
)

const (
	logFieldService  = "service"
	logFieldFile     = "file"
	logFieldOP       = "op"
	logFieldInterval = "interval"

	serviceTypeServer           = "server"
	serviceTypeWatcher          = "watcher"
	serviceTypeSignal           = "signal"
	serviceTypeGarbageCollector = "gc"
)

const (
	// intervalGarbageCollectionOAuth2DPoP is the period between collections of the expired rows of the OAuth2.0 DPoP
	// tables. The rows are small and short lived so the exact period only bounds how many expired rows can accumulate.
	intervalGarbageCollectionOAuth2DPoP = time.Minute * 30
)
