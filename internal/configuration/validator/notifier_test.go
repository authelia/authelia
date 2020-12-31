package validator

import (
	"testing"

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

	suite.Assert().EqualError(suite.validator.Errors()[0], "Notifier should be either `smtp` or `filesystem`")
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

	suite.Assert().EqualError(suite.validator.Errors()[0], "Notifier should be either `smtp` or `filesystem`")
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

// Deprecated: Temporary Test. TODO: Remove in 4.28 (Whole Test).
func (suite *NotifierSuite) TestShouldReturnDeprecationWarningsFor428() {
	var disableVerifyCert = true

	suite.configuration.SMTP.TrustedCert = "/tmp"
	suite.configuration.SMTP.DisableVerifyCert = &disableVerifyCert

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Require().True(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	warnings := suite.validator.Warnings()

	suite.Require().Len(warnings, 2)

	suite.Assert().EqualError(warnings[0], "DEPRECATED: SMTP Notifier `disable_verify_cert` option has been replaced by `notifier.smtp.tls.skip_verify` (will be removed in 4.28.0)")
	suite.Assert().EqualError(warnings[1], "DEPRECATED: SMTP Notifier `trusted_cert` option has been replaced by the global option `certificates_directory` (will be removed in 4.28.0)")

	// Should override since TLS schema is not defined
	suite.Assert().Equal(true, suite.configuration.SMTP.TLS.SkipVerify)
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}
