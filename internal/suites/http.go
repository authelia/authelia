package suites

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewHTTPTransport create a new transport skipping TLS verification and resolving the suite domains itself.
func NewHTTPTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Needs to be enabled in suites. Not used in production.
		},
		DialContext: DialContext,
	}
}

// NewHTTPClient create a new client skipping TLS verification and not redirecting.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: NewHTTPTransport(),
		Timeout:   5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
