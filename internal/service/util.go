package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// RunAll runs the services from every provisioner returned by GetProvisioners.
func RunAll(ctx Context) (err error) {
	provisioners := GetProvisioners()

	return Run(ctx, provisioners...)
}

// Run provisions and runs the services from the given provisioners until the context is cancelled, a SIGINT or
// SIGTERM signal is received, or a service terminates.
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

	rctx := &runContext{base: ctx, Context: cctx}

	for _, provisioner := range provisioners {
		if service, err := provisioner(rctx); err != nil {
			shutdown(log, services)

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

	shutdown(log, services)

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

func shutdown(log *logrus.Entry, services []Provider) {
	wg := &sync.WaitGroup{}

	log.Tracef("Shutdown of %d services is required", len(services))

	for _, service := range services {
		wg.Add(1)

		go func(service Provider) {
			defer wg.Done()

			service.Log().Trace("Shutdown of service initiated")

			service.Shutdown()

			service.Log().Trace("Shutdown of service complete")
		}(service)
	}

	wg.Wait()
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
