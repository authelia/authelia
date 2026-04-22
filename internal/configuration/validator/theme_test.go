package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type Theme struct {
	suite.Suite
	config    *schema.Configuration
	validator *schema.StructValidator
}

func (suite *Theme) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = &schema.Configuration{
		Theme: "light",
	}
}

func (suite *Theme) TestShouldValidateCompleteConfiguration() {
	ValidateTheme(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *Theme) TestShouldRaiseErrorWhenInvalidThemeProvided() {
	suite.config.Theme = testInvalid

	ValidateTheme(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "option 'theme' must be one of 'light', 'dark', 'grey', 'oled', or 'auto' but it's configured as 'invalid'")
}

func (suite *Theme) TestShouldRaiseErrorWhenInvalidCustomCSSProvided() {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldNotValidateOnProtocolRelative", "//example.com/test.css", "option 'custom_css' with value '//example.com/test.css' is invalid: must be an absolute path or an https URL"},
		{"ShouldNotValidateOnHTTP", "http://example.com/test.css", "option 'custom_css' with value 'http://example.com/test.css' is invalid: must be an absolute path or an https URL"},
		{"ShouldNotValidateOnJavascript", "javascript:alert(1)", "option 'custom_css' with value 'javascript:alert(1)' is invalid: must be an absolute path or an https URL"},
		{"ShouldNotValidateOnRelativePath", "custom.css", "option 'custom_css' with value 'custom.css' is invalid: must be an absolute path or an https URL"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			suite.SetupTest()
			suite.config.CustomCSS = tc.have

			ValidateTheme(suite.config, suite.validator)

			suite.Assert().Len(suite.validator.Warnings(), 0)
			suite.Require().Len(suite.validator.Errors(), 1)
			suite.Assert().EqualError(suite.validator.Errors()[0], tc.expected)
		})
	}
}

func (suite *Theme) TestShouldValidateValidCustomCSS() {
	testCases := []struct {
		name string
		have string
	}{
		{"ShouldValidateOnHTTPS", "https://example.com/test.css"},
		{"ShouldValidateOnAbsolutePath", "/custom-assets/styles.css"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			suite.SetupTest()
			suite.config.CustomCSS = tc.have

			ValidateTheme(suite.config, suite.validator)

			suite.Assert().Len(suite.validator.Warnings(), 0)
			suite.Assert().Len(suite.validator.Errors(), 0)
			if tc.have == "https://example.com/test.css" {
				suite.Assert().NotNil(suite.config.CustomCSSURL)
				suite.Assert().Equal("example.com", suite.config.CustomCSSURL.Host)
			}
		})
	}
}

func (suite *Theme) TestShouldWarnOnCSPTemplateIncompatibility() {
	suite.config.CustomCSS = "https://example.com/test.css"
	suite.config.Server.Headers.CSPTemplate = "default-src 'self'"

	ValidateTheme(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 1)
	suite.Assert().EqualError(suite.validator.Warnings()[0], "option 'custom_css' with value 'https://example.com/test.css' appears to have a host which is not explicitly allowed in the 'server.headers.csp_template' which may prevent it from loading")
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *Theme) TestShouldNotWarnOnCSPTemplateCompatibility() {
	suite.config.CustomCSS = "https://example.com/test.css"
	testCases := []string{
		"default-src 'self'; style-src 'self' example.com",
		"default-src 'self'; style-src 'self' https://example.com",
		"default-src 'self'; style-src 'self' *",
		"default-src 'self'; style-src 'self' https://*",
		"style-src example.com",
		"default-src example.com",
	}

	for _, tc := range testCases {
		suite.T().Run(tc, func(t *testing.T) {
			suite.SetupTest()
			suite.config.CustomCSS = "https://example.com/test.css"
			suite.config.Server.Headers.CSPTemplate = schema.CSPTemplate(tc)

			ValidateTheme(suite.config, suite.validator)

			suite.Assert().Len(suite.validator.Warnings(), 0)
			suite.Assert().Len(suite.validator.Errors(), 0)
		})
	}
}

func (suite *Theme) TestShouldWarnWhenHostOnlyInNonStyleDirective() {
	suite.config.CustomCSS = "https://example.com/test.css"
	suite.config.Server.Headers.CSPTemplate = "default-src 'self'; img-src example.com"

	ValidateTheme(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 1)
	suite.Assert().EqualError(suite.validator.Warnings()[0], "option 'custom_css' with value 'https://example.com/test.css' appears to have a host which is not explicitly allowed in the 'server.headers.csp_template' which may prevent it from loading")
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *Theme) TestShouldWarnWhenHostIsSuffixOfAllowedHost() {
	suite.config.CustomCSS = "https://example.com/test.css"
	suite.config.Server.Headers.CSPTemplate = "default-src 'self'; style-src 'self' foo.example.com"

	ValidateTheme(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 1)
	suite.Assert().EqualError(suite.validator.Warnings()[0], "option 'custom_css' with value 'https://example.com/test.css' appears to have a host which is not explicitly allowed in the 'server.headers.csp_template' which may prevent it from loading")
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func TestThemes(t *testing.T) {
	suite.Run(t, new(Theme))
}
