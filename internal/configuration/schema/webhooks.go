// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// Webhooks represents the configuration of the webhook subsystem.
type Webhooks struct {
	StartupCheck bool                 `koanf:"startup_check" yaml:"startup_check" toml:"startup_check" json:"startup_check" jsonschema:"default=false,title=Startup Check" jsonschema_description:"Performs a connectivity check against every destination at startup."`
	Destinations []WebhookDestination `koanf:"destinations" yaml:"destinations,omitempty" toml:"destinations,omitempty" json:"destinations,omitempty" jsonschema:"title=Destinations" jsonschema_description:"The webhook destinations which receive events."`
}

// WebhookDestination represents a single webhook receiver.
type WebhookDestination struct {
	Name             string                `koanf:"name" yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty" jsonschema:"required,title=Name" jsonschema_description:"The unique name of this destination used in logs and metrics."`
	Address          *url.URL              `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"required,format=uri,title=Address" jsonschema_description:"The HTTPS address events are delivered to."`
	Events           []string              `koanf:"events" yaml:"events,omitempty" toml:"events,omitempty" json:"events,omitempty" jsonschema:"required,title=Events" jsonschema_description:"The event types this destination receives, either exact names or prefix globs ending in an asterisk."`
	Timeout          time.Duration         `koanf:"timeout" yaml:"timeout,omitempty" toml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"default=10 seconds,title=Timeout" jsonschema_description:"The timeout for an individual delivery attempt."`
	BufferSize       int                   `koanf:"buffer_size" yaml:"buffer_size,omitempty" toml:"buffer_size,omitempty" json:"buffer_size,omitempty" jsonschema:"default=256,title=Buffer Size" jsonschema_description:"The number of events which may be queued for this destination before events are dropped."`
	DisableRedaction bool                  `koanf:"disable_redaction" yaml:"disable_redaction" toml:"disable_redaction" json:"disable_redaction" jsonschema:"default=false,title=Disable Redaction" jsonschema_description:"Includes credential equivalent values such as one-time codes and reset link URLs in payloads. A receiver holding these values can take over any account."`
	Signature        WebhookSignature      `koanf:"signature" yaml:"signature,omitempty" toml:"signature,omitempty" json:"signature,omitempty" jsonschema:"title=Signature" jsonschema_description:"The HMAC signature properties."`
	Authentication   WebhookAuthentication `koanf:"authentication" yaml:"authentication,omitempty" toml:"authentication,omitempty" json:"authentication,omitempty" jsonschema:"title=Authentication" jsonschema_description:"The credentials presented to the receiver."`
	Headers          map[string]string     `koanf:"headers" yaml:"headers,omitempty" toml:"headers,omitempty" json:"headers,omitempty" jsonschema:"title=Headers" jsonschema_description:"Additional static headers included with every request."`
	Validation       WebhookValidation     `koanf:"validation" yaml:"validation,omitempty" toml:"validation,omitempty" json:"validation,omitempty" jsonschema:"title=Validation" jsonschema_description:"The CloudEvents abuse protection handshake performed before any event is delivered."`
	Batch            WebhookBatch          `koanf:"batch" yaml:"batch,omitempty" toml:"batch,omitempty" json:"batch,omitempty" jsonschema:"title=Batch" jsonschema_description:"The batching properties. Batching is disabled unless 'size' is configured."`
	Retry            WebhookRetry          `koanf:"retry" yaml:"retry,omitempty" toml:"retry,omitempty" json:"retry,omitempty" jsonschema:"title=Retry" jsonschema_description:"The retry policy for failed deliveries."`
	TLS              *TLS                  `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"The TLS connection properties. Disabling certificate verification allows an interceptor to read the credentials and the event bodies sent to this destination, so it should not be used for a destination which is sent credentials."`
}

// WebhookSignature represents the HMAC signature configuration for a destination.
type WebhookSignature struct {
	Secret    string `koanf:"secret" yaml:"secret,omitempty" toml:"secret,omitempty" json:"secret,omitempty" jsonschema:"title=Secret" jsonschema_description:"The shared secret used to sign payloads. Signing is disabled when empty."`
	Algorithm string `koanf:"algorithm" yaml:"algorithm,omitempty" toml:"algorithm,omitempty" json:"algorithm,omitempty" jsonschema:"default=sha256,enum=sha256,enum=sha512,title=Algorithm" jsonschema_description:"The HMAC hash algorithm."`
}

// WebhookAuthentication represents the credentials presented to a destination.
type WebhookAuthentication struct {
	Bearer *WebhookAuthenticationBearer `koanf:"bearer" yaml:"bearer,omitempty" toml:"bearer,omitempty" json:"bearer,omitempty" jsonschema:"title=Bearer" jsonschema_description:"Bearer token authentication."`
	Basic  *WebhookAuthenticationBasic  `koanf:"basic" yaml:"basic,omitempty" toml:"basic,omitempty" json:"basic,omitempty" jsonschema:"title=Basic" jsonschema_description:"Basic authentication."`
}

// WebhookAuthenticationBearer represents bearer token authentication for a destination.
type WebhookAuthenticationBearer struct {
	Token  string `koanf:"token" yaml:"token,omitempty" toml:"token,omitempty" json:"token,omitempty" jsonschema:"required,title=Token" jsonschema_description:"The token value."`
	Header string `koanf:"header" yaml:"header,omitempty" toml:"header,omitempty" json:"header,omitempty" jsonschema:"default=Authorization,title=Header" jsonschema_description:"The header the token is sent in."`
	Scheme string `koanf:"scheme" yaml:"scheme,omitempty" toml:"scheme,omitempty" json:"scheme,omitempty" jsonschema:"default=Bearer,title=Scheme" jsonschema_description:"The scheme prefixed to the token. Configure as empty for a bare API key header."`
}

// WebhookAuthenticationBasic represents basic authentication for a destination.
type WebhookAuthenticationBasic struct {
	Username string `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"required,title=Username" jsonschema_description:"The username."`
	Password string `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"required,title=Password" jsonschema_description:"The password."`
}

// WebhookValidation represents the CloudEvents abuse protection handshake for a destination. Before any event is
// delivered Authelia sends an OPTIONS request to the destination address and requires it to confirm that it accepts
// deliveries from this origin. A destination which does not confirm receives no events.
type WebhookValidation struct {
	Disable     bool   `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Skips the abuse protection handshake and delivers events without asking the destination to confirm it accepts them."`
	Origin      string `koanf:"origin" yaml:"origin,omitempty" toml:"origin,omitempty" json:"origin,omitempty" jsonschema:"title=Origin" jsonschema_description:"The origin presented to the destination. Used only when the origin cannot be derived from the configuration, unless 'force_origin' is enabled."`
	ForceOrigin bool   `koanf:"force_origin" yaml:"force_origin" toml:"force_origin" json:"force_origin" jsonschema:"default=false,title=Force Origin" jsonschema_description:"Presents the configured 'origin' even when an origin can be derived from the configuration."`
	Rate        int    `koanf:"rate" yaml:"rate,omitempty" toml:"rate,omitempty" json:"rate,omitempty" jsonschema:"title=Rate" jsonschema_description:"The delivery rate in requests per minute requested from the destination. An unrestricted rate is requested when this is absent or 0."`
}

// WebhookBatch represents the batching policy for a destination. Batching is disabled unless Size is configured, in
// which case events accumulate until the batch is full or MaxWait elapses and are delivered as a single CloudEvents
// batched content mode request.
type WebhookBatch struct {
	Size      int           `koanf:"size" yaml:"size,omitempty" toml:"size,omitempty" json:"size,omitempty" jsonschema:"title=Size" jsonschema_description:"The number of events delivered in a single request. Batching is disabled when this is absent or 0."`
	MaxWait   time.Duration `koanf:"max_wait" yaml:"max_wait,omitempty" toml:"max_wait,omitempty" json:"max_wait,omitempty" jsonschema:"default=5 seconds,title=Maximum Wait" jsonschema_description:"The longest an event waits for the batch to fill before the batch is delivered regardless."`
	Immediate []string      `koanf:"immediate" yaml:"immediate,omitempty" toml:"immediate,omitempty" json:"immediate,omitempty" jsonschema:"title=Immediate" jsonschema_description:"The event types which bypass batching and are delivered on their own as soon as they occur, either exact names or prefix globs ending in an asterisk."`
}

// WebhookRetry represents the retry policy for a destination.
type WebhookRetry struct {
	Attempts        int           `koanf:"attempts" yaml:"attempts,omitempty" toml:"attempts,omitempty" json:"attempts,omitempty" jsonschema:"default=3,title=Attempts" jsonschema_description:"The maximum number of delivery attempts including the first."`
	InitialInterval time.Duration `koanf:"initial_interval" yaml:"initial_interval,omitempty" toml:"initial_interval,omitempty" json:"initial_interval,omitempty" jsonschema:"default=1 second,title=Initial Interval" jsonschema_description:"The backoff interval before the second attempt."`
	MaximumInterval time.Duration `koanf:"maximum_interval" yaml:"maximum_interval,omitempty" toml:"maximum_interval,omitempty" json:"maximum_interval,omitempty" jsonschema:"default=1 minute,title=Maximum Interval" jsonschema_description:"The ceiling applied to the backoff interval."`
}

// DefaultWebhookDestination represents the default configuration for a webhook destination.
var DefaultWebhookDestination = WebhookDestination{
	Timeout:    time.Second * 10,
	BufferSize: 256,
	Signature: WebhookSignature{
		Algorithm: "sha256",
	},
	Batch: WebhookBatch{
		MaxWait: time.Second * 5,
	},
	Retry: WebhookRetry{
		Attempts:        3,
		InitialInterval: time.Second,
		MaximumInterval: time.Minute,
	},
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}
