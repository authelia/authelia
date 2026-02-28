package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// RunAll the provided Context with all Provisioner's.
func RunAll(ctx Context) (err error) {
	provisioners := GetProvisioners()

	return Run(ctx, provisioners...)
}

// Run the provided Context and Provisioner's.
func Run(ctx Context, provisioners ...Provisioner) (err error) {
	cctx, cancel := context.WithCancel(ctx)

	group, cctx := errgroup.WithContext(cctx)

	defer cancel()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(quit)

	var (
		services []Provider
	)

	log := ctx.GetLogger()

	for _, provisioner := range provisioners {
		if service, err := provisioner(ctx); err != nil {
			return fmt.Errorf("error occurred provisioning services: %w", err)
		} else if service != nil {
			services = append(services, service)
		}
	}

	for _, service := range services {
		group.Go(service.Run)
	}

	log.Info("Startup complete")

	select {
	case s := <-quit:
		log.WithField("signal", s.String()).Debug("Shutdown initiated due to process signal")
	case <-cctx.Done():
		log.Debug("Shutdown initiated due to context completion")
	}

	cancel()

	log.Info("Shutdown initiated")

	wgShutdown := &sync.WaitGroup{}

	log.Tracef("Shutdown of %d services is required", len(services))

	for _, service := range services {
		wgShutdown.Add(1)

		go func(service Provider) {
			service.Log().Trace("Shutdown of service initiated")

			service.Shutdown()

			wgShutdown.Done()

			service.Log().Trace("Shutdown of service complete")
		}(service)
	}

	wgShutdown.Wait()

	if err = ctx.GetProviders().UserProvider.Close(); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred closing authentication connections")
	}

	if err = ctx.GetProviders().StorageProvider.Close(); err != nil {
		log.WithError(err).Error("Error occurred closing database connections")
	}

	if err = group.Wait(); err != nil {
		log.WithError(err).Error("Error occurred waiting for shutdown")
	}

	log.Info("Shutdown complete")

	return nil
}

// IsConfigFileWatcherEnabled determines if a configuration change will cause the Authelia application to
// reload.
func IsConfigFileWatcherEnabled() (enabled bool) {
	if value, err := strconv.ParseBool(os.Getenv("X_AUTHELIA_CONFIG_RELOAD")); err != nil && value {
		return true
	}

	return false
}

func connectionType(isTLS bool) string {
	if isTLS {
		return "TLS"
	}

	return "non-TLS"
}

func recoverErr(i any) error {
	switch v := i.(type) {
	case nil:
		return nil
	case string:
		return fmt.Errorf("recovered panic: %s", v)
	case error:
		return fmt.Errorf("recovered panic: %w", v)
	default:
		return fmt.Errorf("recovered panic with unknown type: %v", v)
	}
}

func newFileWatcherPaths(in []string) (paths FileWatcherPaths, err error) {
	paths = make(FileWatcherPaths, len(in))

	for i, path := range in {
		if path == "" {
			return nil, fmt.Errorf("path must not be empty")
		}

		if path, err = filepath.Abs(path); err != nil {
			return nil, fmt.Errorf("could not determine the absolute path of file '%s': %w", path, err)
		}

		var info os.FileInfo

		if info, err = os.Stat(path); err != nil {
			switch {
			case os.IsNotExist(err):
				return nil, fmt.Errorf("error stating file '%s': file does not exist", path)
			case os.IsPermission(err):
				return nil, fmt.Errorf("error stating file '%s': permission denied trying to read the file", path)
			default:
				return nil, fmt.Errorf("error stating file '%s': %w", path, err)
			}
		}

		if info.IsDir() {
			paths[i] = FileWatcherPath{
				Directory: path,
				Info:      info,
			}
		} else {
			paths[i] = FileWatcherPath{
				File:      filepath.Base(path),
				Directory: filepath.Dir(path),
				Info:      info,
			}
		}
	}

	return paths, nil
}
