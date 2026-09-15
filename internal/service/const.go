// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import "errors"

const (
	fmtLogServerListening = "Listening for %s connections on '%s' path '%s'"
)

const (
	logFieldService   = "service"
	logFieldFile      = "file"
	logFieldOP        = "op"
	logFieldProvider  = "provider"
	logFieldFrequency = "frequency"

	serviceTypeServer  = "server"
	serviceTypeWatcher = "watcher"
	serviceTypeSignal  = "signal"
	serviceTypeGC      = "gc"
)

const (
	// environmentVariableConfigReload enables reloading the application when the configuration changes.
	environmentVariableConfigReload = "X_AUTHELIA_CONFIG_RELOAD"

	// environmentVariableConfigReloadPaths is a comma delimited list of additional paths to watch for changes.
	environmentVariableConfigReloadPaths = "X_AUTHELIA_CONFIG_RELOAD_PATHS"
)

var (
	// ErrApplicationReload is emitted when the application is being reloaded. This effectively starts Authelia again
	// instead of doing a full exit.
	ErrApplicationReload = errors.New("application reload")
)
