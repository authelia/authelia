package services

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func connectionType(isTLS bool) string {
	if isTLS {
		return "TLS"
	}

	return "non-TLS"
}

func Run(ctx context.Context, config *schema.Configuration, providers middlewares.Providers, log *logrus.Logger) {
	cctx, cancel := context.WithCancel(ctx)

	group, cctx := errgroup.WithContext(cctx)

	defer cancel()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(quit)

	var (
		services []Provider
	)

	provisioners := GetProvisioners()

	for _, provisioner := range provisioners {
		if service, err := provisioner(config, providers, log); err != nil {
			service.Log().Trace("Service Loaded")

			services = append(services, service)

			group.Go(service.Run)
		}
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

	var err error

	if err = providers.StorageProvider.Close(); err != nil {
		log.WithError(err).Error("Error occurred closing database connections")
	}

	if err = group.Wait(); err != nil {
		log.WithError(err).Error("Error occurred waiting for shutdown")
	}

	log.Info("Shutdown complete")
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
