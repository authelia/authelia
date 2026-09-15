// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

const (
	fmtLogServerListening = "Listening for %s connections on '%s' path '%s'"
)

const (
	logFieldService   = "service"
	logFieldFile      = "file"
	logFieldOP        = "op"
	logFieldProvider  = "provider"
	logFieldFrequency = "frequency"
	logFieldInterval  = "interval"

	serviceTypeServer   = "server"
	serviceTypeWatcher  = "watcher"
	serviceTypeSignal   = "signal"
	serviceTypeGC       = "gc"
	serviceTypeWatchdog = "watchdog"

	serviceNameSystemd = "systemd"
	serviceNameReload  = "reload"
)

const (
	statusReady    = "Authelia is ready"
	statusStopping = "Authelia is shutting down"
)
