package server

import (
	"embed"
	"net"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

//go:embed public_html
var assets embed.FS

// Start Authelia's internal webserver with the given configuration and providers.
func Start(config schema.Configuration, providers middlewares.Providers) {
	logger := logging.Logger()

	handler := getRequestHandler(config, providers)

	server := &fasthttp.Server{
		ErrorHandler:          handleError,
		Handler:               handler,
		NoDefaultServerHeader: true,
		ReadBufferSize:        config.Server.ReadBufferSize,
		WriteBufferSize:       config.Server.WriteBufferSize,
	}

	address := net.JoinHostPort(config.Server.Host, strconv.Itoa(config.Server.Port))

	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatalf("Error initializing listener: %s", err)
	}

	if config.Server.TLS.Certificate != "" && config.Server.TLS.Key != "" {
		if err = writeHealthCheckEnv(config.Server.DisableHealthcheck, schemeHTTPS, config.Server.Host, config.Server.Path, config.Server.Port); err != nil {
			logger.Fatalf("Could not configure healthcheck: %v", err)
		}

		if config.Server.Path == "" {
			logger.Infof("Listening for TLS connections on '%s' path '/'", address)
		} else {
			logger.Infof("Listening for TLS connections on '%s' paths '/' and '%s'", address, config.Server.Path)
		}

		logger.Fatal(server.ServeTLS(listener, config.Server.TLS.Certificate, config.Server.TLS.Key))
	} else {
		if err = writeHealthCheckEnv(config.Server.DisableHealthcheck, schemeHTTP, config.Server.Host, config.Server.Path, config.Server.Port); err != nil {
			logger.Fatalf("Could not configure healthcheck: %v", err)
		}

		if config.Server.Path == "" {
			logger.Infof("Listening for non-TLS connections on '%s' path '/'", address)
		} else {
			logger.Infof("Listening for non-TLS connections on '%s' paths '/' and '%s'", address, config.Server.Path)
		}
		logger.Fatal(server.Serve(listener))
	}
}
