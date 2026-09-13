// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type WebhooksSuite struct {
	suite.Suite

	config    schema.Webhooks
	validator *schema.StructValidator
}

func (suite *WebhooksSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.Webhooks{
		Destinations: []schema.WebhookDestination{
			{
				Name:    "admin-api",
				Address: &url.URL{Scheme: "https", Host: "admin.example.com", Path: "/hooks"},
				Events:  []string{"user.*"},
			},
		},
	}
}

func (suite *WebhooksSuite) TestShouldApplyDefaults() {
	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
	suite.Equal(time.Second*10, suite.config.Destinations[0].Timeout)
	suite.Equal(256, suite.config.Destinations[0].BufferSize)
	suite.Equal("sha256", suite.config.Destinations[0].Signature.Algorithm)
	suite.Equal(3, suite.config.Destinations[0].Retry.Attempts)
	suite.NotNil(suite.config.Destinations[0].TLS)
}

func (suite *WebhooksSuite) TestShouldRejectNonHTTPSAddress() {
	suite.config.Destinations[0].Address = &url.URL{Scheme: "http", Host: "admin.example.com"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'address' with value 'http://admin.example.com' is invalid: the scheme must be 'https' but it's configured as 'http'")
}

func (suite *WebhooksSuite) TestShouldRejectAddressWithUserInfo() {
	suite.config.Destinations[0].Address = &url.URL{Scheme: "https", Host: "admin.example.com", User: url.UserPassword("john", "abc123")}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Contains(suite.validator.Errors()[0].Error(), "it must not contain user information")
}

func (suite *WebhooksSuite) TestShouldRejectDuplicateNames() {
	suite.config.Destinations = append(suite.config.Destinations, suite.config.Destinations[0])

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'name' must be unique")
}

func (suite *WebhooksSuite) TestShouldRejectBothAuthenticationMethods() {
	suite.config.Destinations[0].Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc"}
	suite.config.Destinations[0].Authentication.Basic = &schema.WebhookAuthenticationBasic{Username: "john", Password: "abc"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: please ensure only one of the 'bearer' or 'basic' authentication methods is configured")
}

func (suite *WebhooksSuite) TestShouldRejectReservedHeader() {
	suite.config.Destinations[0].Headers = map[string]string{"content-type": "text/plain"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the reserved header 'Content-Type'")
}

func (suite *WebhooksSuite) TestShouldRejectHeaderUsedByAuthentication() {
	suite.config.Destinations[0].Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc"}
	suite.config.Destinations[0].Headers = map[string]string{"Authorization": "abc"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the reserved header 'Authorization'")
}

func (suite *WebhooksSuite) TestShouldRejectSelectorMatchingNoEventType() {
	suite.config.Destinations[0].Events = []string{"users.*"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'events' with value 'users.*' is invalid: it doesn't match any known event type")
}

func (suite *WebhooksSuite) TestShouldRejectUnknownSignatureAlgorithm() {
	suite.config.Destinations[0].Signature.Secret = "abc"
	suite.config.Destinations[0].Signature.Algorithm = "md5"

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': signature: option 'algorithm' must be one of 'sha256' or 'sha512' but it's configured as 'md5'")
}

func (suite *WebhooksSuite) TestShouldNotValidateWhenNoDestinations() {
	suite.config.Destinations = nil

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
}

func (suite *WebhooksSuite) TestShouldRejectAnAddressWithNoHost() {
	testCases := []struct {
		name    string
		address string
	}{
		{"ShouldRejectAnEmptyAuthority", "https:///hooks"},
		{"ShouldRejectAPortWithNoHost", "https://:443/hooks"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			address, err := url.Parse(tc.address)

			suite.Require().NoError(err)

			suite.config.Destinations[0].Address = address

			ValidateWebhooks(&suite.config, suite.validator)

			suite.Require().Len(suite.validator.Errors(), 1)
			suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'address' with value '"+tc.address+"' is invalid: it must have a host")
		})
	}
}

func (suite *WebhooksSuite) TestShouldRejectAnInvalidHeaderName() {
	suite.config.Destinations[0].Headers = map[string]string{"X Tenant": "acme"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the header name 'X Tenant' as it's not a valid HTTP header name")
}

func (suite *WebhooksSuite) TestShouldRejectAnInvalidHeaderValue() {
	suite.config.Destinations[0].Headers = map[string]string{"X-Tenant": "acme\nX-Injected: evil"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the value for header 'X-Tenant' as it's not a valid HTTP header value")
}

func (suite *WebhooksSuite) TestShouldRejectAnInvalidBearerHeaderName() {
	suite.config.Destinations[0].Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc", Header: "Api Key"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: bearer: option 'header' with value 'Api Key' is invalid: it's not a valid HTTP header name")
}

func (suite *WebhooksSuite) TestShouldRejectAnInvalidBearerTokenValue() {
	suite.config.Destinations[0].Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc\r\nX-Injected: evil"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: bearer: the assembled value of options 'scheme' and 'token' is not a valid HTTP header value")
}

func (suite *WebhooksSuite) TestShouldAcceptValidHeadersAndBearer() {
	suite.config.Destinations[0].Headers = map[string]string{"X-Tenant": "acme"}
	suite.config.Destinations[0].Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
}

func (suite *WebhooksSuite) TestShouldRejectForcedOriginWithoutAnOrigin() {
	suite.config.Destinations[0].Validation.ForceOrigin = true

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': validation: option 'origin' is required when option 'force_origin' is enabled")
}

func (suite *WebhooksSuite) TestShouldAcceptForcedOriginWithAnOrigin() {
	suite.config.Destinations[0].Validation.ForceOrigin = true
	suite.config.Destinations[0].Validation.Origin = "auth.example.com"

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
}

func (suite *WebhooksSuite) TestShouldRejectANegativeValidationRate() {
	suite.config.Destinations[0].Validation.Rate = -1

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': validation: option 'rate' must not be negative but it's configured as '-1'")
}

func (suite *WebhooksSuite) TestShouldNotValidateValidationOptionsWhenDisabled() {
	suite.config.Destinations[0].Validation.Disable = true
	suite.config.Destinations[0].Validation.ForceOrigin = true
	suite.config.Destinations[0].Validation.Rate = -1

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
}

func (suite *WebhooksSuite) TestShouldNotEnableBatchingByDefault() {
	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
	suite.Equal(0, suite.config.Destinations[0].Batch.Size)
}

func (suite *WebhooksSuite) TestShouldApplyBatchDefaultsWhenEnabled() {
	suite.config.Destinations[0].Batch.Size = 50

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
	suite.Equal(50, suite.config.Destinations[0].Batch.Size)
	suite.Equal(time.Second*5, suite.config.Destinations[0].Batch.MaxWait)
}

func (suite *WebhooksSuite) TestShouldRejectNegativeBatchSize() {
	suite.config.Destinations[0].Batch.Size = -1

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': batch: option 'size' must not be negative but it's configured as '-1'")
}

func (suite *WebhooksSuite) TestShouldRejectBatchSizeGreaterThanBufferSize() {
	suite.config.Destinations[0].BufferSize = 10
	suite.config.Destinations[0].Batch.Size = 11

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': batch: option 'size' must not be greater than option 'buffer_size' which is '10' but it's configured as '11'")
}

func (suite *WebhooksSuite) TestShouldAllowBatchSizeEqualToBufferSize() {
	suite.config.Destinations[0].BufferSize = 10
	suite.config.Destinations[0].Batch.Size = 10

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
}

func (suite *WebhooksSuite) TestShouldAcceptImmediateSelectors() {
	suite.config.Destinations[0].Batch.Size = 50
	suite.config.Destinations[0].Batch.Immediate = []string{"security.*", "user.password.changed"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
}

func (suite *WebhooksSuite) TestShouldRejectImmediateSelectorMatchingNothing() {
	suite.config.Destinations[0].Batch.Size = 50
	suite.config.Destinations[0].Batch.Immediate = []string{"not.a.real.type"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': batch: option 'immediate' with value 'not.a.real.type' is invalid: it doesn't match any known event type")
}

func TestWebhooksSuite(t *testing.T) {
	suite.Run(t, new(WebhooksSuite))
}

func (suite *WebhooksSuite) TestShouldDefaultTheBearerSchemeRegardlessOfHeaderCase() {
	testCases := []string{"Authorization", "authorization", "AUTHORIZATION"}

	for _, header := range testCases {
		suite.Run(header, func() {
			suite.SetupTest()

			suite.config.Destinations[0].Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc", Header: header}

			ValidateWebhooks(&suite.config, suite.validator)

			suite.Len(suite.validator.Errors(), 0)
			suite.Equal("Bearer", suite.config.Destinations[0].Authentication.Bearer.Scheme)
		})
	}
}

func (suite *WebhooksSuite) TestShouldNotDefaultTheBearerSchemeForACustomHeader() {
	suite.config.Destinations[0].Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc", Header: "X-API-Key"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
	suite.Equal("", suite.config.Destinations[0].Authentication.Bearer.Scheme)
}
