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

func TestWebhooksSuite(t *testing.T) {
	suite.Run(t, new(WebhooksSuite))
}

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
				Events:  []string{"com.authelia.user.*"},
				Authentication: schema.WebhookAuthentication{
					Bearer: &schema.WebhookAuthenticationBearer{Token: "abc"},
				},
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
	suite.Equal("Bearer", suite.config.Destinations[0].Authentication.Bearer.Scheme)
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

func (suite *WebhooksSuite) TestShouldRejectBearerAndBasicAuthentication() {
	suite.config.Destinations[0].Authentication.Basic = &schema.WebhookAuthenticationBasic{Username: "john", Password: "abc"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: please ensure only one of the 'bearer' or 'basic' options is configured")
}

func (suite *WebhooksSuite) TestShouldAcceptBasicAuthentication() {
	suite.config.Destinations[0].Authentication.Bearer = nil
	suite.config.Destinations[0].Authentication.Basic = &schema.WebhookAuthenticationBasic{Username: "john", Password: "abc"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
}

func (suite *WebhooksSuite) TestShouldRejectIncompleteBasicAuthentication() {
	suite.config.Destinations[0].Authentication.Bearer = nil
	suite.config.Destinations[0].Authentication.Basic = &schema.WebhookAuthenticationBasic{}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 2)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: basic: option 'username' is required")
	suite.EqualError(suite.validator.Errors()[1], "webhooks: destinations: destination 'admin-api': authentication: basic: option 'password' is required")
}

func (suite *WebhooksSuite) TestShouldRejectMissingAuthentication() {
	suite.config.Destinations[0].Authentication.Bearer = nil

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: one of the 'bearer' or 'basic' options is required")
}

func (suite *WebhooksSuite) TestShouldRejectAnAddressCarryingTheAccessToken() {
	suite.config.Destinations[0].Address.RawQuery = "tenant=acme&access_token=abc"

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'address' with value 'https://admin.example.com/hooks' is invalid: it must not contain the 'access_token' query parameter, configure the token with the 'authentication.bearer' options instead")
}

func (suite *WebhooksSuite) TestShouldNotPrintTheAccessTokenInOtherAddressErrors() {
	suite.config.Destinations[0].Address = &url.URL{Scheme: "http", Host: "admin.example.com", Path: "/hooks", RawQuery: "access_token=abc"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 2)

	for _, err := range suite.validator.Errors() {
		suite.NotContains(err.Error(), "abc")
	}
}

func (suite *WebhooksSuite) TestShouldDefaultTheBearerMethod() {
	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
	suite.Equal("header", suite.config.Destinations[0].Authentication.Bearer.Method)
}

func (suite *WebhooksSuite) TestShouldAcceptTheQueryBearerMethod() {
	suite.config.Destinations[0].Authentication.Bearer.Method = "query"

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
	suite.Equal("", suite.config.Destinations[0].Authentication.Bearer.Scheme)
}

func (suite *WebhooksSuite) TestShouldRejectASchemeWithTheQueryBearerMethod() {
	suite.config.Destinations[0].Authentication.Bearer.Method = "query"
	suite.config.Destinations[0].Authentication.Bearer.Scheme = "Bearer"

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: bearer: option 'scheme' must not be configured when option 'method' is 'query'")
}

func (suite *WebhooksSuite) TestShouldRejectAnUnknownBearerMethod() {
	suite.config.Destinations[0].Authentication.Bearer.Method = "cookie"

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': authentication: bearer: option 'method' must be one of 'header' or 'query' but it's configured as 'cookie'")
}

func (suite *WebhooksSuite) TestShouldValidateTheValidationOrigin() {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldAcceptADNSName", "auth.example.com", ""},
		{"ShouldAcceptAnUppercaseDNSName", "Auth.Example.com", ""},
		{"ShouldRejectAURL", "https://auth.example.com", "webhooks: destinations: destination 'admin-api': validation: option 'origin' with value 'https://auth.example.com' is invalid: it must be a DNS name"},
		{"ShouldRejectAnAsterisk", "*", "webhooks: destinations: destination 'admin-api': validation: option 'origin' with value '*' is invalid: it must be a DNS name"},
		{"ShouldRejectAHostWithAPort", "auth.example.com:443", "webhooks: destinations: destination 'admin-api': validation: option 'origin' with value 'auth.example.com:443' is invalid: it must be a DNS name"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			suite.config.Destinations[0].Validation.Origin = tc.have

			ValidateWebhooks(&suite.config, suite.validator)

			if tc.expected == "" {
				suite.Len(suite.validator.Errors(), 0)

				return
			}

			suite.Require().Len(suite.validator.Errors(), 1)
			suite.EqualError(suite.validator.Errors()[0], tc.expected)
		})
	}
}

func (suite *WebhooksSuite) TestShouldPreserveAConfiguredScheme() {
	suite.config.Destinations[0].Authentication.Bearer.Scheme = "Token"

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Len(suite.validator.Errors(), 0)
	suite.Equal("Token", suite.config.Destinations[0].Authentication.Bearer.Scheme)
}

func (suite *WebhooksSuite) TestShouldRejectReservedHeader() {
	suite.config.Destinations[0].Headers = map[string]string{"content-type": "text/plain"}

	ValidateWebhooks(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the reserved header 'Content-Type'")
}

func (suite *WebhooksSuite) TestShouldRejectCloudEventsHeaders() {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldRejectLowerCase", "ce-id", "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the reserved header 'Ce-Id'"},
		{"ShouldRejectUpperCase", "CE-SOURCE", "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the reserved header 'Ce-Source'"},
		{"ShouldRejectCanonicalCase", "Ce-Time", "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the reserved header 'Ce-Time'"},
		{"ShouldRejectAnUnknownAttribute", "ce-tenant", "webhooks: destinations: destination 'admin-api': option 'headers' must not contain the reserved header 'Ce-Tenant'"},
		{"ShouldAcceptANameWithoutTheHyphen", "Cell", ""},
		{"ShouldAcceptANameContainingThePrefix", "X-Ce-Id", ""},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			suite.config.Destinations[0].Headers = map[string]string{tc.have: "acme"}

			ValidateWebhooks(&suite.config, suite.validator)

			if tc.expected == "" {
				suite.Len(suite.validator.Errors(), 0)

				return
			}

			suite.Require().Len(suite.validator.Errors(), 1)
			suite.EqualError(suite.validator.Errors()[0], tc.expected)
		})
	}
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
	suite.config.Destinations[0].Batch.Immediate = []string{"com.authelia.security.*", "com.authelia.user.password.changed"}

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
