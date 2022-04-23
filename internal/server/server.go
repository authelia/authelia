package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// CreateServer Create Authelia's internal webserver with the given configuration and providers.
func CreateServer(config schema.Configuration, providers middlewares.Providers) (server *fasthttp.Server, listener net.Listener) {
	logger := logging.Logger()

	handler := getHandler(config, providers)

	server = &fasthttp.Server{
		ErrorHandler:          handlerError(),
		Handler:               handler,
		NoDefaultServerHeader: true,
		ReadBufferSize:        config.Server.ReadBufferSize,
		WriteBufferSize:       config.Server.WriteBufferSize,
	}

	address := net.JoinHostPort(config.Server.Host, strconv.Itoa(config.Server.Port))

	var (
		err              error
		connectionType   string
		connectionScheme string
	)

	if config.Server.TLS.Certificate != "" && config.Server.TLS.Key != "" {
		connectionType, connectionScheme = "TLS", schemeHTTPS

		if err = server.AppendCert(config.Server.TLS.Certificate, config.Server.TLS.Key); err != nil {
			logger.Fatalf("unable to load certificate: %v", err)
		}

		if len(config.Server.TLS.ClientCertificates) > 0 {
			caCertPool := x509.NewCertPool()

			for _, path := range config.Server.TLS.ClientCertificates {
				cert, err := os.ReadFile(path)
				if err != nil {
					logger.Fatalf("Cannot read client TLS certificate %s: %s", path, err)
				}

				caCertPool.AppendCertsFromPEM(cert)
			}

			// ClientCAs should never be nil, otherwise the system cert pool is used for client authentication
			// but we don't want everybody on the Internet to be able to authenticate.
			server.TLSConfig.ClientCAs = caCertPool
			server.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		if listener, err = tls.Listen("tcp", address, server.TLSConfig.Clone()); err != nil {
			logger.Fatalf("Error initializing listener: %s", err)
		}
	} else {
		connectionType, connectionScheme = "non-TLS", schemeHTTP

		if listener, err = net.Listen("tcp", address); err != nil {
			logger.Fatalf("Error initializing listener: %s", err)
		}
	}

	if err = writeHealthCheckEnv(config.Server.DisableHealthcheck, connectionScheme, config.Server.Host,
		config.Server.Path, config.Server.Port); err != nil {
		logger.Fatalf("Could not configure healthcheck: %v", err)
	}

	if config.Server.Path == "" {
		logger.Infof("Initializing server for %s connections on '%s' path '/'", connectionType, listener.Addr().String())
	} else {
		logger.Infof("Initializing server for %s connections on '%s' paths '/' and '%s'", connectionType, listener.Addr().String(), config.Server.Path)
	}

	return server, listener
}

// CreateMetricsServer creates a metrics server.
func CreateMetricsServer(config schema.TelemetryMetricsConfig) (server *fasthttp.Server, listener net.Listener, err error) {
	if listener, err = config.Address.Listener(); err != nil {
		return nil, nil, err
	}

	server = &fasthttp.Server{
		ErrorHandler:          handlerError(),
		NoDefaultServerHeader: true,
		Handler:               getMetricsHandler(),
	}

	return server, listener, nil
}
