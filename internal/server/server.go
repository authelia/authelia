package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// CreateDefaultServer Create Authelia's internal webserver with the given configuration and providers.
func CreateDefaultServer(config schema.Configuration, providers middlewares.Providers) (server *fasthttp.Server, listener net.Listener, err error) {
	server = &fasthttp.Server{
		ErrorHandler:          handleError(),
		Handler:               handleRouter(config, providers),
		NoDefaultServerHeader: true,
		ReadBufferSize:        config.Server.ReadBufferSize,
		WriteBufferSize:       config.Server.WriteBufferSize,
	}

	address := net.JoinHostPort(config.Server.Host, strconv.Itoa(config.Server.Port))

	var (
		connectionType   string
		connectionScheme string
	)

	if config.Server.TLS.Certificate != "" && config.Server.TLS.Key != "" {
		connectionType, connectionScheme = "TLS", schemeHTTPS

		if err = server.AppendCert(config.Server.TLS.Certificate, config.Server.TLS.Key); err != nil {
			return nil, nil, fmt.Errorf("unable to load tls server certificate '%s' or private key '%s': %w", config.Server.TLS.Certificate, config.Server.TLS.Key, err)
		}

		if len(config.Server.TLS.ClientCertificates) > 0 {
			caCertPool := x509.NewCertPool()

			var cert []byte

			for _, path := range config.Server.TLS.ClientCertificates {
				if cert, err = os.ReadFile(path); err != nil {
					return nil, nil, fmt.Errorf("unable to load tls client certificate '%s': %w", path, err)
				}

				caCertPool.AppendCertsFromPEM(cert)
			}

			// ClientCAs should never be nil, otherwise the system cert pool is used for client authentication
			// but we don't want everybody on the Internet to be able to authenticate.
			server.TLSConfig.ClientCAs = caCertPool
			server.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		if listener, err = tls.Listen("tcp", address, server.TLSConfig.Clone()); err != nil {
			return nil, nil, fmt.Errorf("unable to initialize tcp listener: %w", err)
		}
	} else {
		connectionType, connectionScheme = "non-TLS", schemeHTTP

		if listener, err = net.Listen("tcp", address); err != nil {
			return nil, nil, fmt.Errorf("unable to initialize tcp listener: %w", err)
		}
	}

	if err = writeHealthCheckEnv(config.Server.DisableHealthcheck, connectionScheme, config.Server.Host,
		config.Server.Path, config.Server.Port); err != nil {
		return nil, nil, fmt.Errorf("unable to configure healthcheck: %w", err)
	}

	logger := logging.Logger()

	if config.Server.Path == "" {
		logger.Infof("Initializing server for %s connections on '%s' path '/'", connectionType, listener.Addr().String())
	} else {
		logger.Infof("Initializing server for %s connections on '%s' paths '/' and '%s'", connectionType, listener.Addr().String(), config.Server.Path)
	}

	return server, listener, nil
}

// CreateMetricsServer creates a metrics server.
func CreateMetricsServer(config schema.TelemetryMetricsConfig) (server *fasthttp.Server, listener net.Listener, err error) {
	if listener, err = config.Address.Listener(); err != nil {
		return nil, nil, err
	}

	server = &fasthttp.Server{
		ErrorHandler:          handleError(),
		NoDefaultServerHeader: true,
		Handler:               handleMetrics(),
	}

	return server, listener, nil
}
