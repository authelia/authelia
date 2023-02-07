package commands

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

type ServiceRun func() error

type ServerServiceFactory func(*schema.Configuration, middlewares.Providers) (*fasthttp.Server, net.Listener, error)

type ReloadFilter func(path string) (skipped bool)

type ProviderReload interface {
	Reload() (reloaded bool, err error)
}

type Service interface {
	Run() (err error)
	Shutdown()
}

type ServerService struct {
	server   *fasthttp.Server
	listener net.Listener
	log      *logrus.Entry
}

func NewServerService(name string, server *fasthttp.Server, listener net.Listener, logger *logrus.Logger) (service *ServerService) {
	entry := logger.WithFields(logrus.Fields{
		"service": "server",
		"server":  name,
	})

	return &ServerService{
		server:   server,
		listener: listener,
		log:      entry,
	}
}

func (service *ServerService) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	if err = service.server.Serve(service.listener); err != nil {
		service.log.WithError(err).Error("Error returned attempting to serve requests")

		return err
	}

	return nil
}

func (service *ServerService) Shutdown() {
	if err := service.server.Shutdown(); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}
}

type FileWatcherService struct {
	watcher *fsnotify.Watcher
	reload  ProviderReload

	log       *logrus.Entry
	file      string
	directory string
}

func NewFileWatcherService(name, path string, reload ProviderReload, logger *logrus.Logger) (service *FileWatcherService, err error) {
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

	entry := logger.WithFields(logrus.Fields{
		"service": "watcher",
		"watcher": name,
	})

	if info.IsDir() {
		service = &FileWatcherService{
			watcher:   watcher,
			reload:    reload,
			log:       entry,
			directory: filepath.Clean(path),
		}
	} else {
		service = &FileWatcherService{
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

func (service *FileWatcherService) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	for {
		select {
		case event, ok := <-service.watcher.Events:
			if !ok {
				return nil
			}

			if service.file != "" && service.file != filepath.Base(event.Name) {
				service.log.WithField("file", event.Name).WithField("op", event.Op).Tracef("File modification detected to irrelevant file")
				break
			}

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write, event.Op&fsnotify.Create == fsnotify.Create:
				service.log.WithField("file", event.Name).WithField("op", event.Op).Debug("File modification was detected")

				var reloaded bool

				switch reloaded, err = service.reload.Reload(); {
				case err != nil:
					service.log.WithField("file", event.Name).WithField("op", event.Op).WithError(err).Error("Error occurred during reload")
				case reloaded:
					service.log.WithField("file", event.Name).Info("Reloaded successfully")
				default:
					service.log.WithField("file", event.Name).Debug("Reload of was triggered but it was skipped")
				}
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				service.log.WithField("file", event.Name).WithField("op", event.Op).Debug("File remove was detected")
			}
		case err, ok := <-service.watcher.Errors:
			if !ok {
				return nil
			}

			service.log.WithError(err).Errorf("Error while watching files")
		}
	}
}

func (service *FileWatcherService) Shutdown() {
	if err := service.watcher.Close(); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}
}
