// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpguts"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
)

// ValidateWebhooks validates and updates the webhooks configuration.
func ValidateWebhooks(config *schema.Webhooks, validator *schema.StructValidator) {
	if len(config.Destinations) == 0 {
		return
	}

	names := map[string]struct{}{}

	for i := range config.Destinations {
		validateWebhookDestination(i, &config.Destinations[i], names, validator)
	}
}

func validateWebhookDestination(i int, config *schema.WebhookDestination, names map[string]struct{}, validator *schema.StructValidator) {
	if config.Name == "" {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationNameRequired, i+1))

		return
	}

	if _, ok := names[config.Name]; ok {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationNameDuplicate, config.Name))

		return
	}

	names[config.Name] = struct{}{}

	validateWebhookDestinationAddress(config, validator)
	validateWebhookDestinationEvents(config, validator)
	validateWebhookDestinationSignature(config, validator)
	validateWebhookDestinationAuthentication(config, validator)
	validateWebhookDestinationHeaders(config, validator)
	validateWebhookDestinationDefaults(config, validator)
	validateWebhookDestinationBatch(config, validator)
	validateWebhookDestinationValidation(config, validator)
}

func validateWebhookDestinationAddress(config *schema.WebhookDestination, validator *schema.StructValidator) {
	if config.Address == nil {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAddressRequired, config.Name))

		return
	}

	if config.Address.Scheme != "https" {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAddressScheme, config.Name, webhookAddressString(config.Address), config.Address.Scheme))
	}

	if config.Address.Hostname() == "" {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAddressHost, config.Name, webhookAddressString(config.Address)))
	}

	if config.Address.User != nil {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAddressUserInfo, config.Name, webhookAddressString(config.Address)))
	}

	if config.Address.Fragment != "" {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAddressFragment, config.Name, webhookAddressString(config.Address)))
	}

	if config.Address.Query().Has(schema.WebhookQueryParameterAccessToken) {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAddressAccessToken, config.Name, webhookAddressString(config.Address)))
	}
}

func webhookAddressString(address *url.URL) string {
	if !address.Query().Has(schema.WebhookQueryParameterAccessToken) {
		return address.String()
	}

	redacted := *address

	redacted.RawQuery = ""

	return redacted.String()
}

func validateWebhookDestinationEvents(config *schema.WebhookDestination, validator *schema.StructValidator) {
	if len(config.Events) == 0 {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationEventsRequired, config.Name))

		return
	}

	for _, selector := range config.Events {
		if len(events.Match(selector)) == 0 {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationEventsUnknown, config.Name, selector))
		}
	}
}

func validateWebhookDestinationSignature(config *schema.WebhookDestination, validator *schema.StructValidator) {
	if config.Signature.Secret == "" {
		return
	}

	if config.Signature.Algorithm == "" {
		config.Signature.Algorithm = schema.DefaultWebhookDestination.Signature.Algorithm
	}

	switch config.Signature.Algorithm {
	case events.SignatureAlgorithmSHA256, events.SignatureAlgorithmSHA512:
		break
	default:
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationSignatureAlg, config.Name, config.Signature.Algorithm))
	}
}

func validateWebhookDestinationAuthentication(config *schema.WebhookDestination, validator *schema.StructValidator) {
	switch {
	case config.Authentication.Bearer == nil && config.Authentication.Basic == nil:
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthRequired, config.Name))

		return
	case config.Authentication.Bearer != nil && config.Authentication.Basic != nil:
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthMultiple, config.Name))

		return
	}

	if config.Authentication.Basic != nil {
		if config.Authentication.Basic.Username == "" {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthBasicOption, config.Name, "username"))
		}

		if config.Authentication.Basic.Password == "" {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthBasicOption, config.Name, "password"))
		}

		return
	}

	if config.Authentication.Bearer.Token == "" {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthBearerToken, config.Name))
	}

	switch config.Authentication.Bearer.Method {
	case "":
		config.Authentication.Bearer.Method = schema.WebhookAuthenticationMethodHeader
	case schema.WebhookAuthenticationMethodHeader:
		break
	case schema.WebhookAuthenticationMethodQuery:
		if config.Authentication.Bearer.Scheme != "" {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthBearerSchemeQuery, config.Name))
		}

		return
	default:
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthBearerMethod, config.Name, config.Authentication.Bearer.Method))

		return
	}

	if config.Authentication.Bearer.Scheme == "" {
		config.Authentication.Bearer.Scheme = schema.DefaultWebhookAuthenticationScheme
	}

	if !httpguts.ValidHeaderFieldValue(config.Authentication.Bearer.Scheme + " " + config.Authentication.Bearer.Token) {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationAuthBearerValue, config.Name))
	}
}

func validateWebhookDestinationHeaders(config *schema.WebhookDestination, validator *schema.StructValidator) {
	if len(config.Headers) == 0 {
		return
	}

	reserved := map[string]struct{}{
		"Authorization": {},
		"Cache-Control": {},
		"Content-Type":  {},
		"User-Agent":    {},
	}

	for name, value := range config.Headers {
		if !httpguts.ValidHeaderFieldName(name) {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationHeaderName, config.Name, name))

			continue
		}

		if !httpguts.ValidHeaderFieldValue(value) {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationHeaderValue, config.Name, name))
		}

		canonical := http.CanonicalHeaderKey(name)

		if _, ok := reserved[canonical]; ok {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationHeaderReserved, config.Name, canonical))

			continue
		}

		lower := strings.ToLower(canonical)

		if strings.HasPrefix(lower, "x-authelia-") || strings.HasPrefix(lower, events.HeaderPrefixCloudEvents) {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationHeaderReserved, config.Name, canonical))
		}
	}
}

func validateWebhookDestinationValidation(config *schema.WebhookDestination, validator *schema.StructValidator) {
	if config.Validation.Disable {
		return
	}

	if config.Validation.ForceOrigin && config.Validation.Origin == "" {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationValidationOrigin, config.Name))
	}

	if config.Validation.Origin != "" && !reDomainCharacters.MatchString(strings.ToLower(config.Validation.Origin)) {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationValidationOriginInvalid, config.Name, config.Validation.Origin))
	}

	if config.Validation.Rate < 0 {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationValidationRate, config.Name, config.Validation.Rate))
	}
}

func validateWebhookDestinationBatch(config *schema.WebhookDestination, validator *schema.StructValidator) {
	if config.Batch.Size < 0 {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationBatchSize, config.Name, config.Batch.Size))

		return
	}

	if config.Batch.Size == 0 {
		return
	}

	if config.Batch.Size > config.BufferSize {
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationBatchSizeBuffer, config.Name, config.BufferSize, config.Batch.Size))
	}

	if config.Batch.MaxWait <= 0 {
		config.Batch.MaxWait = schema.DefaultWebhookDestination.Batch.MaxWait
	}

	for _, selector := range config.Batch.Immediate {
		if len(events.Match(selector)) == 0 {
			validator.Push(fmt.Errorf(errFmtWebhooksDestinationBatchImmediate, config.Name, selector))
		}
	}
}

func validateWebhookDestinationDefaults(config *schema.WebhookDestination, validator *schema.StructValidator) {
	if config.Timeout <= 0 {
		config.Timeout = schema.DefaultWebhookDestination.Timeout
	}

	switch {
	case config.BufferSize == 0:
		config.BufferSize = schema.DefaultWebhookDestination.BufferSize
	case config.BufferSize < 0:
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationBufferSize, config.Name, config.BufferSize))
	}

	switch {
	case config.Retry.Attempts == 0:
		config.Retry.Attempts = schema.DefaultWebhookDestination.Retry.Attempts
	case config.Retry.Attempts < 0:
		validator.Push(fmt.Errorf(errFmtWebhooksDestinationRetryAttempts, config.Name, config.Retry.Attempts))
	}

	if config.Retry.InitialInterval <= 0 {
		config.Retry.InitialInterval = schema.DefaultWebhookDestination.Retry.InitialInterval
	}

	if config.Retry.MaximumInterval <= 0 {
		config.Retry.MaximumInterval = schema.DefaultWebhookDestination.Retry.MaximumInterval
	}

	if config.Signature.Algorithm == "" {
		config.Signature.Algorithm = schema.DefaultWebhookDestination.Signature.Algorithm
	}

	if config.TLS == nil {
		config.TLS = &schema.TLS{MinimumVersion: schema.DefaultWebhookDestination.TLS.MinimumVersion}
	}
}
