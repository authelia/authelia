package suites

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewHTTPClient create a new client skipping TLS verification and not redirecting.
func NewHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Needs to be enabled in suites. Not used in production.
		},
	}

	return &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
