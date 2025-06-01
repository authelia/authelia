package service

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/server"
)

// ProvisionServer initializes the main listener.
func ProvisionServer(ctx Context) (service Provider, err error) {
	var (
		s        *fasthttp.Server
		listener net.Listener
		paths    []string
		isTLS    bool
	)

	switch s, listener, paths, isTLS, err = server.New(ctx.GetConfiguration(), ctx.GetProviders()); {
	case err != nil:
		return nil, err
	case s != nil && listener != nil:
		service = NewBaseServer("main", s, listener, paths, isTLS, ctx.GetLogger())
	default:
		return nil, nil
	}

	return service, nil
}

// ProvisionServerMetrics initializes the metrics listener.
func ProvisionServerMetrics(ctx Context) (service Provider, err error) {
	var (
		s        *fasthttp.Server
		listener net.Listener
		paths    []string
		isTLS    bool
	)

	switch s, listener, paths, isTLS, err = server.NewMetrics(ctx.GetConfiguration(), ctx.GetProviders()); {
	case err != nil:
		return nil, err
	case s != nil && listener != nil:
		service = NewBaseServer("metrics", s, listener, paths, isTLS, ctx.GetLogger())
	default:
		return nil, nil
	}

	return service, nil
}

// NewBaseServer creates a new Server with the appropriate logger etc.
func NewBaseServer(name string, server *fasthttp.Server, listener net.Listener, paths []string, isTLS bool, log *logrus.Entry) (service *Server) {
	return &Server{
		name:     name,
		server:   server,
		listener: listener,
		paths:    paths,
		isTLS:    isTLS,
		log:      log.WithFields(map[string]any{logFieldService: serviceTypeServer, serviceTypeServer: name}),
	}
}

// Server is a Provider which runs a web server.
type Server struct {
	name     string
	server   *fasthttp.Server
	paths    []string
	isTLS    bool
	listener net.Listener
	log      *logrus.Entry
}

// ServiceType returns the service type for this service, which is always 'server'.
func (service *Server) ServiceType() string {
	return serviceTypeServer
}

// ServiceName returns the individual name for this service.
func (service *Server) ServiceName() string {
	return service.name
}

// Run the Server.
func (service *Server) Run() (err error) {
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

// Shutdown the Server.
func (service *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	if err := service.server.ShutdownWithContext(ctx); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}
}

// Log returns the *logrus.Entry of the Server.
func (service *Server) Log() *logrus.Entry {
	return service.log
}
