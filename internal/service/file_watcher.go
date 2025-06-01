// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ProvisionConfigFileWatcher returns a Provider which watches the configuration for changes and reloads the
// application when a change occurs. It's only provisioned when the relevant environment variable is enabled.
func ProvisionConfigFileWatcher(ctx Context) (service Provider, err error) {
	if !IsConfigFileWatcherEnabled() {
		return nil, nil
	}

	// The context paths are copied as they must not be mutated by the append below.
	paths := slices.Clone(ctx.GetConfigurationPaths())

	if additional := utils.StringSplitClean(os.Getenv(environmentVariableConfigReloadPaths), ","); len(additional) != 0 {
		paths = append(paths, additional...)
	}

	if len(paths) == 0 {
		ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeWatcher, serviceTypeWatcher: "configuration"}).
			Warn("Configuration reloading was enabled but no configuration paths are available to watch")

		return nil, nil
	}

	action := func(log *logrus.Entry, event fsnotify.Event) (bubble bool, err error) {
		if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) {
			log.Info("Configuration change detected, reloading the application")

			return true, ErrApplicationReload
		}

		return false, nil
	}

	if service, err = NewFileWatcher("configuration", nil, action, ctx.GetLogger(), paths...); err != nil {
		return nil, err
	}

	return service, nil
}

// ProvisionUsersFileWatcher returns a Provider which watches the file based user database for changes.
func ProvisionUsersFileWatcher(ctx Context) (service Provider, err error) {
	config := ctx.GetConfiguration()
	providers := ctx.GetProviders()

	if config.AuthenticationBackend.File != nil && config.AuthenticationBackend.File.Watch {
		provider, ok := providers.UserProvider.(*authentication.FileUserProvider)

		if !ok {
			return nil, errors.New("error occurred asserting user provider")
		}

		if service, err = NewFileWatcher("users", provider, nil, ctx.GetLogger(), config.AuthenticationBackend.File.Path); err != nil {
			return nil, err
		}
	}

	return service, nil
}

// NewFileWatcher creates a new FileWatcher with the appropriate logger etc. Either a ReloadableProvider or a
// FileWatcherAction must be provided to handle the relevant events.
func NewFileWatcher(name string, reload ReloadableProvider, action FileWatcherAction, log *logrus.Entry, paths ...string) (service *FileWatcher, err error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("error initializing file watcher: path must be specified")
	}

	var fwp FileWatcherPaths

	if fwp, err = newFileWatcherPaths(paths); err != nil {
		return nil, fmt.Errorf("error initializing file watcher: %w", err)
	}

	var watcher *fsnotify.Watcher

	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}

	entry := log.WithFields(map[string]any{logFieldService: serviceTypeWatcher, serviceTypeWatcher: name})

	service = &FileWatcher{
		name:    name,
		watcher: watcher,
		reload:  reload,
		action:  action,
		log:     entry,
		paths:   fwp,
	}

	for _, path := range fwp {
		if err = service.watcher.Add(path.Directory); err != nil {
			if errClose := service.watcher.Close(); errClose != nil {
				entry.WithError(errClose).Error("Error occurred closing the file watcher")
			}

			return nil, fmt.Errorf("failed to add path '%s' to watch list: %w", path.Directory, err)
		}
	}

	return service, nil
}

// FileWatcher is a Provider that watches files for changes.
type FileWatcher struct {
	name string

	watcher *fsnotify.Watcher

	reload ReloadableProvider
	action FileWatcherAction

	log   *logrus.Entry
	paths FileWatcherPaths
}

// FileWatcherPath describes a path being checked by a FileWatcher.
type FileWatcherPath struct {
	File      string
	Directory string
	Info      os.FileInfo
}

// FileWatcherPaths is a composite type that describes a slice of FileWatcherPath.
type FileWatcherPaths []FileWatcherPath

// IsMatch returns true if a fsnotify.Event matches any FileWatcherPath.
func (fwp FileWatcherPaths) IsMatch(event fsnotify.Event) (match bool) {
	directory, file := filepath.Dir(event.Name), filepath.Base(event.Name)

	for _, path := range fwp {
		if directory != path.Directory {
			continue
		}

		if path.Info.IsDir() || file == path.File {
			return true
		}
	}

	return false
}

// ServiceType returns the service type for this service, which is always 'watcher'.
func (service *FileWatcher) ServiceType() string {
	return serviceTypeWatcher
}

// ServiceName returns the individual name for this service.
func (service *FileWatcher) ServiceName() string {
	return service.name
}

// Run the FileWatcher.
func (service *FileWatcher) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	for _, path := range service.paths {
		service.log.WithField(logFieldFile, filepath.Join(path.Directory, path.File)).Info("Watching file for changes")
	}

	for {
		select {
		case event, ok := <-service.watcher.Events:
			if !ok {
				return nil
			}

			if err = service.handleEvent(event); err != nil {
				return err
			}
		case errWatch, ok := <-service.watcher.Errors:
			if !ok {
				return nil
			}

			service.handleError(errWatch)
		}
	}
}

// handleEvent handles a single fsnotify.Event. A non-nil error indicates the FileWatcher should terminate and the
// error should be returned to the caller.
func (service *FileWatcher) handleEvent(event fsnotify.Event) (err error) {
	log := service.log.WithFields(map[string]any{logFieldFile: event.Name, logFieldOP: event.Op})

	if !service.paths.IsMatch(event) {
		log.Trace("File modification detected to irrelevant file")

		return nil
	}

	switch {
	case service.reload != nil:
		service.handleEventReload(log, event)
	case service.action != nil:
		var bubble bool

		switch bubble, err = service.action(log, event); {
		case err != nil && bubble:
			return err
		case err != nil:
			log.WithError(err).Error("Error occurred during action")
		default:
			log.Debug("Action triggered successfully")
		}
	default:
		log.Debug("File event was detected")
	}

	return nil
}

func (service *FileWatcher) handleEventReload(log *logrus.Entry, event fsnotify.Event) {
	switch {
	case event.Op&fsnotify.Write == fsnotify.Write, event.Op&fsnotify.Create == fsnotify.Create:
		log.Debug("File modification was detected")

		var (
			reloaded bool
			err      error
		)

		switch reloaded, err = service.reload.Reload(); {
		case err != nil:
			var e errWatcher
			if errors.As(err, &e) && !e.WatcherReloadErrorCritical() {
				log.WithError(err).Debug("Reload was triggered but it was skipped")
			} else {
				log.WithError(err).Error("Error occurred during reload")
			}
		case reloaded:
			log.Info("Reloaded successfully")
		default:
			log.Debug("Reload was triggered but it was skipped")
		}
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		log.Debug("File remove was detected")
	}
}

func (service *FileWatcher) handleError(err error) {
	service.log.WithError(err).Error("Error while watching file for changes")
}

// Shutdown the FileWatcher.
func (service *FileWatcher) Shutdown() {
	if err := service.watcher.Close(); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}
}

// Log returns the *logrus.Entry of the FileWatcher.
func (service *FileWatcher) Log() *logrus.Entry {
	return service.log
}
