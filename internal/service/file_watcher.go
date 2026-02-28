package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
)

// ProvisionConfigFileWatcher creates a new FileWatcher that checks for configuration changes.
func ProvisionConfigFileWatcher(ctx Context) (service Provider, err error) {
	enabled := IsConfigFileWatcherEnabled()

	if enabled {
		action := func(log *logrus.Entry, event fsnotify.Event) (bubble bool, err error) {
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) {
				return true, ErrApplicationReload
			}

			return false, nil
		}

		paths := ctx.GetConfigurationPaths()

		if additional := utils.StringSplitClean(os.Getenv("X_AUTHELIA_CONFIG_RELOAD_PATHS"), ","); len(additional) > 0 {
			paths = append(paths, additional...)
		}

		if service, err = NewFileWatcher("configuration", nil, action, ctx.GetLogger(), paths...); err != nil {
			return nil, err
		}
	}

	return service, nil
}

// ProvisionUsersFileWatcher creates a new FileWatcher that checks for user database changes.
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

// NewFileWatcher creates a new FileWatcher with the appropriate logger etc.
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
func (fwp *FileWatcherPaths) IsMatch(event fsnotify.Event) (match bool) {
	directory, file := filepath.Dir(event.Name), filepath.Base(event.Name)

	for _, path := range *fwp {
		if path.Info.IsDir() {
			return directory == path.Directory
		}

		if directory != path.Directory {
			continue
		}

		if file != path.File {
			continue
		}

		return true
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
//
//nolint:gocyclo
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

			log := service.log.WithFields(map[string]any{logFieldFile: event.Name, logFieldOP: event.Op})

			if !service.paths.IsMatch(event) {
				log.Trace("File modification detected to irrelevant file")

				break
			}

			switch {
			case service.reload != nil:
				switch {
				case event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename):
					log.Debug("File modification was detected")

					var reloaded bool

					switch reloaded, err = service.reload.Reload(); {
					case err != nil:
						log.WithError(err).Error("Error occurred during reload")
					case reloaded:
						log.Info("Reloaded successfully")
					default:
						log.Debug("Reload was triggered but it was skipped")
					}
				case event.Has(fsnotify.Remove):
					log.Debug("File remove was detected")
				}
			case service.action != nil:
				var bubble bool

				if bubble, err = service.action(log, event); err != nil {
					if bubble {
						return err
					}

					log.WithError(err).Error("Error occurred during action")
				} else {
					log.Debug("Action triggered successfully")
				}
			default:
				log.Debug("File event was detected")
			}
		case err, ok := <-service.watcher.Errors:
			if !ok {
				return nil
			}

			service.log.WithError(err).Error("Error while watching file for changes")
		}
	}
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
