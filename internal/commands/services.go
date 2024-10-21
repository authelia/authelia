package commands

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/server"
)

// NewServerService creates a new ServerService with the appropriate logger etc.
func NewServerService(name string, server *fasthttp.Server, listener net.Listener, paths []string, isTLS bool, log *logrus.Logger) (service *ServerService) {
	return &ServerService{
		name:     name,
		server:   server,
		listener: listener,
		paths:    paths,
		isTLS:    isTLS,
		log:      log.WithFields(map[string]any{logFieldService: serviceTypeServer, serviceTypeServer: name}),
	}
}

// NewFileWatcherService creates a new FileWatcherService with the appropriate logger etc.
func NewFileWatcherService(name, path string, reload ProviderReload, log *logrus.Logger) (service *FileWatcherService, err error) {
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
		service = &FileWatcherService{
			name:      name,
			watcher:   watcher,
			reload:    reload,
			log:       entry,
			directory: filepath.Clean(path),
		}
	} else {
		service = &FileWatcherService{
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

// NewSignalService creates a new SignalService with the appropriate logger etc.
func NewSignalService(name string, action func() (err error), log *logrus.Logger, signals ...os.Signal) (service *SignalService) {
	return &SignalService{
		name:    name,
		signals: signals,
		action:  action,
		log:     log.WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: name}),
	}
}

type ServiceCtx interface {
	GetLogger() *logrus.Logger
	GetProviders() middlewares.Providers
	GetConfiguration() *schema.Configuration

	context.Context
}

// ProviderReload represents the required methods to support reloading a provider.
type ProviderReload interface {
	Reload() (reloaded bool, err error)
}

// Service represents the required methods to support handling a service.
type Service interface {
	// ServiceType returns the type name for the Service.
	ServiceType() string

	// ServiceName returns the individual name for the Service.
	ServiceName() string

	// Run performs the running operations for the Service.
	Run() (err error)

	// Shutdown perform the shutdown cleanup and termination operations for the Service.
	Shutdown()

	// Log returns the logger configured for the service.
	Log() *logrus.Entry
}

// ServerService is a Service which runs a web server.
type ServerService struct {
	name     string
	server   *fasthttp.Server
	paths    []string
	isTLS    bool
	listener net.Listener
	log      *logrus.Entry
}

// ServiceType returns the service type for this service, which is always 'server'.
func (service *ServerService) ServiceType() string {
	return serviceTypeServer
}

// ServiceName returns the individual name for this service.
func (service *ServerService) ServiceName() string {
	return service.name
}

// Run the ServerService.
func (service *ServerService) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	service.log.Infof(fmtLogServerListening, connectionType(service.isTLS), service.listener.Addr().String(), strings.Join(service.paths, "' and '"))

	if err = service.server.Serve(service.listener); err != nil {
		service.log.WithError(err).Error("Error returned attempting to serve requests")

		return err
	}

	return nil
}

// Shutdown the ServerService.
func (service *ServerService) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	if err := service.server.ShutdownWithContext(ctx); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}
}

// Log returns the *logrus.Entry of the ServerService.
func (service *ServerService) Log() *logrus.Entry {
	return service.log
}

// FileWatcherService is a Service that watches files for changes.
type FileWatcherService struct {
	name string

	watcher *fsnotify.Watcher
	reload  ProviderReload

	log       *logrus.Entry
	file      string
	directory string
}

// ServiceType returns the service type for this service, which is always 'watcher'.
func (service *FileWatcherService) ServiceType() string {
	return serviceTypeWatcher
}

// ServiceName returns the individual name for this service.
func (service *FileWatcherService) ServiceName() string {
	return service.name
}

// Run the FileWatcherService.
func (service *FileWatcherService) Run() (err error) {
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

// Shutdown the FileWatcherService.
func (service *FileWatcherService) Shutdown() {
	if err := service.watcher.Close(); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}
}

// Log returns the *logrus.Entry of the FileWatcherService.
func (service *FileWatcherService) Log() *logrus.Entry {
	return service.log
}

// SignalService is a Service which performs actions on signals.
type SignalService struct {
	name    string
	signals []os.Signal
	action  func() (err error)
	log     *logrus.Entry

	notify chan os.Signal
	quit   chan struct{}
}

// ServiceType returns the service type for this service, which is always 'server'.
func (service *SignalService) ServiceType() string {
	return serviceTypeSignal
}

// ServiceName returns the individual name for this service.
func (service *SignalService) ServiceName() string {
	return service.name
}

// Run the ServerService.
func (service *SignalService) Run() (err error) {
	service.quit = make(chan struct{})

	service.notify = make(chan os.Signal, 1)

	signal.Notify(service.notify, service.signals...)

	for {
		select {
		case s := <-service.notify:
			if err = service.action(); err != nil {
				service.log.WithError(err).Error("Error occurred executing service action.")
			} else {
				service.log.WithFields(map[string]any{"signal-received": s.String()}).Debug("Successfully executed service action.")
			}
		case <-service.quit:
			return
		}
	}
}

// Shutdown the ServerService.
func (service *SignalService) Shutdown() {
	signal.Stop(service.notify)

	service.quit <- struct{}{}
}

// Log returns the *logrus.Entry of the ServerService.
func (service *SignalService) Log() *logrus.Entry {
	return service.log
}

func svcSvrMainFunc(ctx ServiceCtx) (service Service) {
	switch svr, listener, paths, isTLS, err := server.CreateDefaultServer(ctx.GetConfiguration(), ctx.GetProviders()); {
	case err != nil:
		ctx.GetLogger().WithError(err).Fatal("Create Server Service (main) returned error")
	case svr != nil && listener != nil:
		service = NewServerService("main", svr, listener, paths, isTLS, ctx.GetLogger())
	default:
		ctx.GetLogger().Fatal("Create Server Service (main) failed")
	}

	return service
}

func svcSvrMetricsFunc(ctx ServiceCtx) (service Service) {
	switch svr, listener, paths, isTLS, err := server.CreateMetricsServer(ctx.GetConfiguration(), ctx.GetProviders()); {
	case err != nil:
		ctx.GetLogger().WithError(err).Fatal("Create Server Service (metrics) returned error")
	case svr != nil && listener != nil:
		service = NewServerService("metrics", svr, listener, paths, isTLS, ctx.GetLogger())
	default:
		ctx.GetLogger().Debug("Create Server Service (metrics) skipped")
	}

	return service
}

func svcWatcherUsersFunc(ctx ServiceCtx) (service Service) {
	var err error

	config := ctx.GetConfiguration()

	if config.AuthenticationBackend.File != nil && config.AuthenticationBackend.File.Watch {
		provider := ctx.GetProviders().UserProvider.(*authentication.FileUserProvider)

		if service, err = NewFileWatcherService("users", config.AuthenticationBackend.File.Path, provider, ctx.GetLogger()); err != nil {
			ctx.GetLogger().WithError(err).Fatal("Create Watcher Service (users) returned error")
		}
	}

	return service
}

func svcSignalLogReOpenFunc(ctx ServiceCtx) (service Service) {
	config := ctx.GetConfiguration()

	if config.Log.FilePath == "" {
		return nil
	}

	return NewSignalService("log-reload", logging.Reopen, ctx.GetLogger(), syscall.SIGHUP)
}

func connectionType(isTLS bool) string {
	if isTLS {
		return "TLS"
	}

	return "non-TLS"
}

func servicesRun(ctx ServiceCtx) {
	cctx, cancel := context.WithCancel(ctx)

	group, cctx := errgroup.WithContext(cctx)

	defer cancel()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(quit)

	var (
		services []Service
	)

	for _, serviceFunc := range []func(ctx ServiceCtx) Service{
		svcSvrMainFunc, svcSvrMetricsFunc,
		svcWatcherUsersFunc, svcSignalLogReOpenFunc,
	} {
		if service := serviceFunc(ctx); service != nil {
			service.Log().Trace("Service Loaded")

			services = append(services, service)

			group.Go(service.Run)
		}
	}

	ctx.GetLogger().Info("Startup complete")

	select {
	case s := <-quit:
		ctx.GetLogger().WithField("signal", s.String()).Debug("Shutdown initiated due to process signal")
	case <-cctx.Done():
		ctx.GetLogger().Debug("Shutdown initiated due to context completion")
	}

	cancel()

	ctx.GetLogger().Info("Shutdown initiated")

	wgShutdown := &sync.WaitGroup{}

	ctx.GetLogger().Tracef("Shutdown of %d services is required", len(services))

	for _, service := range services {
		wgShutdown.Add(1)

		go func(service Service) {
			service.Log().Trace("Shutdown of service initiated")

			service.Shutdown()

			wgShutdown.Done()

			service.Log().Trace("Shutdown of service complete")
		}(service)
	}

	wgShutdown.Wait()

	var err error

	if err = ctx.GetProviders().StorageProvider.Close(); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred closing database connections")
	}

	if err = group.Wait(); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred waiting for shutdown")
	}

	ctx.GetLogger().Info("Shutdown complete")
}
