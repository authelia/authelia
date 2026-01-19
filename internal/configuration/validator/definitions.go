package validator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func ValidateDefinitions(config *schema.Configuration, validator *schema.StructValidator) {
	for name := range config.Definitions.UserAttributes {
		if !isUserAttributeDefinitionNameValid(name, config) {
			validator.Push(fmt.Errorf(errFmtDefinitionsUserAttributesReservedOrDefined, name, name))
		}
	}

	validateWebhookDefinitions(config.Definitions.Webhooks, validator)
}

func validateWebhookDefinitions(webhooks map[string]schema.Webhook, validator *schema.StructValidator) {
	for name, webhook := range webhooks {
		if validated, ok := validateWebhookDefinition(name, webhook, validator); ok {
			webhooks[name] = validated
		}
	}
}

func validateWebhookDefinition(name string, webhook schema.Webhook, validator *schema.StructValidator) (schema.Webhook, bool) {
	// Validate URL is present.
	if webhook.URL == "" {
		validator.Push(fmt.Errorf("definitions: webhooks: %s: option 'url' is required", name))

		return webhook, false
	}

	// Parse and validate URL.
	parsedURL, err := url.Parse(webhook.URL)
	if err != nil {
		validator.Push(fmt.Errorf("definitions: webhooks: %s: option 'url' is invalid: %w", name, err))

		return webhook, false
	}

	// Validate URL has a host.
	if parsedURL.Host == "" {
		validator.Push(fmt.Errorf("definitions: webhooks: %s: option 'url' must include a hostname", name))

		return webhook, false
	}

	// Enforce HTTPS for security.
	if parsedURL.Scheme != "https" {
		validator.Push(fmt.Errorf("definitions: webhooks: %s: option 'url' must use HTTPS for security, got: %s", name, parsedURL.Scheme))

		return webhook, false
	}

	// Validate and normalize HTTP method.
	if webhook.Method == "" {
		webhook.Method = schema.DefaultWebhookConfiguration.Method
	} else {
		method := strings.ToUpper(webhook.Method)

		if method != "POST" && method != "PUT" && method != "PATCH" {
			validator.Push(fmt.Errorf("definitions: webhooks: %s: option 'method' must be POST, PUT, or PATCH, got: %s", name, webhook.Method))

			return webhook, false
		}

		// Normalize method to uppercase.
		webhook.Method = method
	}

	// Set default timeout if not specified.
	if webhook.Timeout == 0 {
		webhook.Timeout = schema.DefaultWebhookConfiguration.Timeout
	}

	// Validate TLS config if provided.
	if webhook.TLS != nil {
		configDefaultTLS := &schema.TLS{
			ServerName: parsedURL.Hostname(),
		}

		if err := ValidateTLSConfig(webhook.TLS, configDefaultTLS); err != nil {
			validator.Push(fmt.Errorf("definitions: webhooks: %s: tls configuration is invalid: %w", name, err))

			return webhook, false
		}
	}

	return webhook, true
}
