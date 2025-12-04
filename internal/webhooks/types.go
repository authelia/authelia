// Package webhooks provides reusable HTTP webhook client infrastructure for Authelia.
// This package can be used by any subsystem that needs to send HTTP webhooks,
// such as notifications, audit logging, or telemetry.
package webhooks

import (
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Config represents the configuration for a webhook endpoint.
// It contains all the parameters needed to make HTTP webhook requests.
type Config struct {
	// URL is the webhook endpoint URL where HTTP requests will be sent.
	// Must use HTTPS scheme for security.
	URL string

	// Method is the HTTP method to use for webhook requests.
	// Supported values: POST, PUT, PATCH
	Method string

	// Headers contains custom HTTP headers to include in webhook requests.
	// Useful for authentication (e.g., Authorization: Bearer token123)
	Headers map[string]string

	// Timeout is the maximum duration to wait for the webhook request to complete.
	// Includes connection time, request transmission, and response receipt.
	Timeout time.Duration

	// TLS contains optional TLS configuration for the webhook connection.
	// Allows customizing server name, skip verify, and TLS version constraints.
	TLS *schema.TLS
}
