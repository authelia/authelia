package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func ProvisionUsersFileWatcher(config *schema.Configuration, providers middlewares.Providers, log *logrus.Logger) (service Provider, err error) {
	if config.AuthenticationBackend.File != nil && config.AuthenticationBackend.File.Watch {
		provider, ok := providers.UserProvider.(*authentication.FileUserProvider)

		if !ok {
			return nil, errors.New("error occurred asserting user provider")
		}

		if service, err = NewFileWatcher("users", config.AuthenticationBackend.File.Path, provider, log); err != nil {
			return nil, err
		}
	}

	return service, nil
}

// NewFileWatcher creates a new FileWatcher with the appropriate logger etc.
func NewFileWatcher(name, path string, reload ReloadableProvider, log *logrus.Logger) (service *FileWatcher, err error) {
	if path == "" {
		return nil, fmt.Errorf("path must be specified")
	}

	var info os.FileInfo

	if info, err = os.Stat(path); err != nil {
		return nil, fmt.Errorf("error stating file '%s': %w", path, err)
	}

	if path, err = filepath.Abs(path); err != nil {
		return nil, fmt.Errorf("error determining absolute path of file '%s': %w", path, err)
	}

	var watcher *fsnotify.Watcher

	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}

	entry := log.WithFields(map[string]any{logFieldService: serviceTypeWatcher, serviceTypeWatcher: name})

	if info.IsDir() {
		service = &FileWatcher{
			name:      name,
			watcher:   watcher,
			reload:    reload,
			log:       entry,
			directory: filepath.Clean(path),
		}
	} else {
		service = &FileWatcher{
			name:      name,
			watcher:   watcher,
			reload:    reload,
			log:       entry,
			directory: filepath.Dir(path),
			file:      filepath.Base(path),
		}
	}

	if err = service.watcher.Add(service.directory); err != nil {
		return nil, fmt.Errorf("failed to add path '%s' to watch list: %w", path, err)
	}

	return service, nil
}

// FileWatcher is a Provider that watches files for changes.
type FileWatcher struct {
	name string

	watcher *fsnotify.Watcher
	reload  ReloadableProvider

	log       *logrus.Entry
	file      string
	directory string
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

	service.log.WithField(logFieldFile, filepath.Join(service.directory, service.file)).Info("Watching file for changes")

	for {
		select {
		case event, ok := <-service.watcher.Events:
			if !ok {
				return nil
			}

			log := service.log.WithFields(map[string]any{logFieldFile: event.Name, logFieldOP: event.Op})

			if service.file != "" && service.file != filepath.Base(event.Name) {
				log.Trace("File modification detected to irrelevant file")
				break
			}

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write, event.Op&fsnotify.Create == fsnotify.Create:
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
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				log.Debug("File remove was detected")
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
