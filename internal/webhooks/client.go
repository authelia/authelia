package webhooks

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Client is an HTTP client for sending webhook requests.
// It provides a reusable, thread-safe way to send HTTP webhooks with
// consistent configuration including TLS settings, headers, and timeouts.
type Client struct {
	// url is the webhook endpoint URL
	url string

	// method is the HTTP method (POST, PUT, PATCH)
	method string

	// headers contains custom HTTP headers to include in every request
	headers map[string]string

	// timeout is the request timeout duration
	timeout time.Duration

	// client is the underlying HTTP client with TLS configuration
	client *http.Client
}

// NewClient creates a new webhook client from the provided configuration.
// It initializes an HTTP client with TLS configuration, custom headers, and timeout settings.
// The certPool parameter provides the root CA certificates for TLS verification.
// Returns a thread-safe Client that can be used concurrently to send webhook requests.
func NewClient(config Config, certPool *x509.CertPool) *Client {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    certPool,
	}

	if config.TLS != nil {
		if config.TLS.ServerName != "" {
			tlsConfig.ServerName = config.TLS.ServerName
		}

		if config.TLS.SkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}

		if config.TLS.MinimumVersion != (schema.TLSVersion{}) {
			tlsConfig.MinVersion = config.TLS.MinimumVersion.Value
		}

		if config.TLS.MaximumVersion != (schema.TLSVersion{}) {
			tlsConfig.MaxVersion = config.TLS.MaximumVersion.Value
		}
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &Client{
		url:     config.URL,
		method:  strings.ToUpper(config.Method),
		headers: config.Headers,
		timeout: config.Timeout,
		client:  httpClient,
	}
}

// Send sends a webhook request with the provided payload.
// It creates an HTTP request with the configured method, URL, and headers,
// sends the payload as JSON, and verifies the response status code.
// The context parameter allows for request cancellation and deadline control.
// Returns nil on success (2xx status codes), or an error with response details on failure.
func (c *Client) Send(ctx context.Context, payload []byte) (err error) {
	// Create HTTP request with context for proper cancellation and deadline propagation.
	req, err := http.NewRequestWithContext(ctx, c.method, c.url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set default Content-Type.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Authelia-Webhook-Client")

	// Add custom headers.
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Send request.
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}

	defer func() {
		// Read and discard response body to enable HTTP connection reuse.
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	// Check response status - accept 200 OK, 201 Created, 202 Accepted per webhook standard.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read up to 1KB of response body for error context.
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		bodyStr := string(bodyBytes)

		if bodyStr != "" {
			return fmt.Errorf("webhook returned non-success status: %d %s, body: %s", resp.StatusCode, resp.Status, bodyStr)
		}

		return fmt.Errorf("webhook returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// ValidateConfig validates a webhook configuration at startup.
// It performs comprehensive validation including URL presence and format,
// HTTPS scheme enforcement, and HTTP method verification.
// Returns an error if the configuration is invalid, nil if valid.
// This should be called during application initialization before creating clients.
func ValidateConfig(config Config) (err error) {
	// Validate URL is present.
	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	// Parse and validate URL.
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Enforce HTTPS for security.
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS for security, got: %s", parsedURL.Scheme)
	}

	// Validate HTTP method.
	method := strings.ToUpper(config.Method)
	if method != "POST" && method != "PUT" && method != "PATCH" {
		return fmt.Errorf("unsupported HTTP method: %s (allowed: POST, PUT, PATCH)", method)
	}

	return nil
}
