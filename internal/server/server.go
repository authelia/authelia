package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// CreateDefaultServer Create Authelia's internal web server with the given configuration and providers.
func CreateDefaultServer(config *schema.Configuration, providers middlewares.Providers) (server *fasthttp.Server, listener net.Listener, paths []string, isTLS bool, err error) {
	if err = providers.Templates.LoadTemplatedAssets(assets); err != nil {
		return nil, nil, nil, false, fmt.Errorf("failed to load templated assets: %w", err)
	}

	server = &fasthttp.Server{
		ErrorHandler:          handleError("server"),
		Handler:               handleRouter(config, providers),
		NoDefaultServerHeader: true,
		ReadBufferSize:        config.Server.Buffers.Read,
		WriteBufferSize:       config.Server.Buffers.Write,
		ReadTimeout:           config.Server.Timeouts.Read,
		WriteTimeout:          config.Server.Timeouts.Write,
		IdleTimeout:           config.Server.Timeouts.Idle,
		Logger:                logging.LoggerPrintf(logrus.DebugLevel),
	}

	var (
		connectionScheme = schemeHTTP
	)

	if listener, err = config.Server.Address.Listener(); err != nil {
		return nil, nil, nil, false, fmt.Errorf("error occurred while attempting to initialize main server listener for address '%s': %w", config.Server.Address.String(), err)
	}

	if config.Server.TLS.Certificate != "" && config.Server.TLS.Key != "" {
		isTLS, connectionScheme = true, schemeHTTPS

		if err = server.AppendCert(config.Server.TLS.Certificate, config.Server.TLS.Key); err != nil {
			return nil, nil, nil, false, fmt.Errorf("unable to load tls server certificate '%s' or private key '%s': %w", config.Server.TLS.Certificate, config.Server.TLS.Key, err)
		}

		if len(config.Server.TLS.ClientCertificates) > 0 {
			caCertPool := x509.NewCertPool()

			var cert []byte

			for _, path := range config.Server.TLS.ClientCertificates {
				if cert, err = os.ReadFile(path); err != nil {
					return nil, nil, nil, false, fmt.Errorf("unable to load tls client certificate '%s': %w", path, err)
				}

				caCertPool.AppendCertsFromPEM(cert)
			}

			// ClientCAs should never be nil, otherwise the system cert pool is used for client authentication
			// but we don't want everybody on the Internet to be able to authenticate.
			server.TLSConfig.ClientCAs = caCertPool
			server.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		listener = tls.NewListener(listener, server.TLSConfig.Clone())
	}

	if err = writeHealthCheckEnv(config.Server.DisableHealthcheck, connectionScheme, config.Server.Address.Hostname(),
		config.Server.Address.RouterPath(), config.Server.Address.Port()); err != nil {
		return nil, nil, nil, false, fmt.Errorf("unable to configure healthcheck: %w", err)
	}

	paths = []string{"/"}

	switch config.Server.Address.RouterPath() {
	case "/", "":
		break
	default:
		paths = append(paths, config.Server.Address.RouterPath())
	}

	return server, listener, paths, isTLS, nil
}

// CreateMetricsServer creates a metrics server.
func CreateMetricsServer(config *schema.Configuration, providers middlewares.Providers) (server *fasthttp.Server, listener net.Listener, paths []string, tls bool, err error) {
	if providers.Metrics == nil {
		return
	}

	server = &fasthttp.Server{
		ErrorHandler:          handleError("telemetry.metrics"),
		NoDefaultServerHeader: true,
		Handler:               handleMetrics(config.Telemetry.Metrics.Address.RouterPath()),
		ReadBufferSize:        config.Telemetry.Metrics.Buffers.Read,
		WriteBufferSize:       config.Telemetry.Metrics.Buffers.Write,
		ReadTimeout:           config.Telemetry.Metrics.Timeouts.Read,
		WriteTimeout:          config.Telemetry.Metrics.Timeouts.Write,
		IdleTimeout:           config.Telemetry.Metrics.Timeouts.Idle,
		Logger:                logging.LoggerPrintf(logrus.DebugLevel),
	}

	if listener, err = config.Telemetry.Metrics.Address.Listener(); err != nil {
		return nil, nil, nil, false, fmt.Errorf("error occurred while attempting to initialize metrics telemetry server listener for address '%s': %w", config.Telemetry.Metrics.Address.String(), err)
	}

	return server, listener, []string{config.Telemetry.Metrics.Address.RouterPath()}, false, nil
}
