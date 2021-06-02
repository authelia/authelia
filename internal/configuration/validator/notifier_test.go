package validator

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type NotifierSuite struct {
	suite.Suite
	configuration schema.NotifierConfiguration
	validator     *schema.StructValidator
}

func (suite *NotifierSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration.SMTP = &schema.SMTPNotifierConfiguration{
		Username: "john",
		Password: "password",
		Sender:   "admin@example.com",
		Host:     "example.com",
		Port:     25,
	}
}

func (suite *NotifierSuite) TestShouldEnsureAtLeastSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.configuration.SMTP = nil

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please configure one of the `notifier` providers (`smtp`, `filesystem`, or `plugin`)")
}

func (suite *NotifierSuite) TestShouldEnsureEitherSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasErrors())

	suite.configuration.FileSystem = &schema.FileSystemNotifierConfiguration{
		Filename: "test",
	}

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please do not configure more than one of the `notifer` providers (`smtp`, `filesystem`, or `plugin`)")
}

func (suite *NotifierSuite) TestShouldEnsureFilenameOfFilesystemNotifierIsProvided() {
	suite.configuration.SMTP = nil
	suite.configuration.FileSystem = &schema.FileSystemNotifierConfiguration{
		Filename: "test",
	}
	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.configuration.FileSystem.Filename = ""

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Filename of filesystem notifier must not be empty")
}

func (suite *NotifierSuite) TestShouldEnsureHostAndPortOfSMTPNotifierAreProvided() {
	suite.configuration.FileSystem = nil
	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.configuration.SMTP.Host = ""
	suite.configuration.SMTP.Port = 0

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().True(suite.validator.HasErrors())

	errors := suite.validator.Errors()

	suite.Require().Len(errors, 2)

	suite.Assert().EqualError(errors[0], "Host of SMTP notifier must be provided")
	suite.Assert().EqualError(errors[1], "Port of SMTP notifier must be provided")
}

func (suite *NotifierSuite) TestShouldEnsureSenderOfSMTPNotifierAreProvided() {
	suite.configuration.FileSystem = nil

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.configuration.SMTP.Sender = ""

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Sender of SMTP notifier must be provided")
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}

func TestShouldRaiseErrorWhenNotifierPluginConfiguredOnInvalidOS(t *testing.T) {
	validator := schema.NewStructValidator()
	notifierConfig := schema.NotifierConfiguration{}
	notifierConfig.Plugin = &schema.PluginConfiguration{}

	ValidateNotifier(&notifierConfig, validator)

	if runtime.GOOS == linux || runtime.GOOS == freebsd || runtime.GOOS == darwin {
		require.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "The `notifier` plugin provider name must be set")
	} else {
		require.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "The `notifier` plugin provider is only available on linux, freebsd, and darwin operating systems")
	}
}
