// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/authelia/authelia/v4/internal/dynamic"
)

// ProvisionDynamicConfigFileWatcher returns a Provider which watches the dynamic configuration files for changes and
// hot reloads them when a change occurs. It's only provisioned when the dynamic configuration paths are configured and
// the watch option is enabled.
func ProvisionDynamicConfigFileWatcher(ctx Context) (service Provider, err error) {
	config := ctx.GetConfiguration()
	manager := ctx.GetProviders().DynamicConfiguration

	if config == nil || manager == nil || len(config.DynamicConfiguration.Paths) == 0 || !config.DynamicConfiguration.Watch {
		return nil, nil
	}

	log := ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeWatcher, serviceTypeWatcher: "dynamic-configuration"})

	throttle := dynamic.NewThrottle(dynamic.DefaultThrottleInterval, manager.Reload, log)

	var watcher *FileWatcher

	if watcher, err = NewFileWatcher("dynamic-configuration", throttle, nil, ctx.GetLogger(), manager.Paths()...); err != nil {
		return nil, err
	}

	return &dynamicFileWatcher{FileWatcher: watcher, throttle: throttle}, nil
}

// dynamicFileWatcher is a FileWatcher which also stops its Throttle on shutdown so a pending trailing reload doesn't
// run after the service has stopped.
type dynamicFileWatcher struct {
	*FileWatcher

	throttle *dynamic.Throttle
}

// Shutdown the dynamicFileWatcher.
func (service *dynamicFileWatcher) Shutdown() {
	service.throttle.Stop()

	service.FileWatcher.Shutdown()
}
