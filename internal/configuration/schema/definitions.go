package schema

import (
	"net"
	"time"
)

// Definitions contains reusable configuration definitions that can be referenced
// throughout the Authelia configuration. This promotes configuration reusability
// and consistency across multiple subsystems.
type Definitions struct {
	// Network defines CIDR ranges that can be referenced in access control rules
	Network map[string][]*net.IPNet `koanf:"network" yaml:"network,omitempty" toml:"network,omitempty" json:"network,omitempty" jsonschema:"title=Network Definitions" jsonschema_description:"Networks CIDR ranges that can be utilized elsewhere in the configuration."`

	// UserAttributes defines custom user attribute expressions using CEL
	UserAttributes map[string]UserAttribute `koanf:"user_attributes" yaml:"user_attributes,omitempty" toml:"user_attributes,omitempty" json:"user_attributes,omitempty" jsonschema:"title=User Attributes" jsonschema_description:"User attributes derived from other attributes."`

	// Webhooks defines reusable webhook endpoints for notifications, audit logging, and telemetry
	Webhooks map[string]Webhook `koanf:"webhooks" yaml:"webhooks,omitempty" toml:"webhooks,omitempty" json:"webhooks,omitempty" jsonschema:"title=Webhook Definitions" jsonschema_description:"Webhook endpoints that can be referenced by other components for sending HTTP notifications."`
}

// UserAttribute defines a custom user attribute derived from an expression.
// It allows creating computed attributes based on other user properties.
type UserAttribute struct {
	// Expression is the CEL expression used to compute the attribute value
	Expression string `koanf:"expression" yaml:"expression,omitempty" toml:"expression,omitempty" json:"expression,omitempty" jsonschema:"required,title=Expression" jsonschema_description:"Expression to derive the user attribute using the common expression language."`
}

// Webhook represents a reusable webhook endpoint configuration.
// Webhooks defined here can be referenced by multiple Authelia subsystems
// such as notifications, audit logging, or telemetry to promote configuration reusability.
type Webhook struct {
	// URL is the webhook endpoint URL (must use HTTPS)
	URL string `koanf:"url" yaml:"url" toml:"url" json:"url" jsonschema:"required,title=URL,format=uri" jsonschema_description:"The webhook endpoint URL where requests will be sent."`

	// Method is the HTTP method to use (POST, PUT, or PATCH)
	Method string `koanf:"method" yaml:"method,omitempty" toml:"method,omitempty" json:"method,omitempty" jsonschema:"title=Method,enum=POST,enum=PUT,enum=PATCH" jsonschema_description:"The HTTP method to use when sending webhook requests."`

	// Headers contains custom HTTP headers for authentication or metadata
	Headers map[string]string `koanf:"headers" yaml:"headers,omitempty" toml:"headers,omitempty" json:"headers,omitempty" jsonschema:"title=Headers" jsonschema_description:"Custom HTTP headers to include in webhook requests."`

	// Timeout is the maximum duration to wait for the webhook request
	Timeout time.Duration `koanf:"timeout" yaml:"timeout,omitempty" toml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"title=Timeout,default=5 seconds" jsonschema_description:"The timeout for the webhook HTTP request."`

	// TLS contains optional TLS connection settings
	TLS *TLS `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"TLS configuration for the webhook connection."`
}
