package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
	suite.configuration.FileSystem = nil
}

/*
	Common Tests.
*/
func (suite *NotifierSuite) TestShouldEnsureAtLeastSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.configuration.SMTP = nil

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierNotConfigured)
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

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierMultipleConfigured)
}

/*
	SMTP Tests.
*/
func (suite *NotifierSuite) TestSMTPShouldSetTLSDefaults() {
	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("example.com", suite.configuration.SMTP.TLS.ServerName)
	suite.Assert().Equal("TLS1.2", suite.configuration.SMTP.TLS.MinimumVersion)
	suite.Assert().False(suite.configuration.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldDefaultTLSServerNameToHost() {
	suite.configuration.SMTP.Host = "google.com"
	suite.configuration.SMTP.TLS = &schema.TLSConfig{
		MinimumVersion: "TLS1.1",
	}

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("google.com", suite.configuration.SMTP.TLS.ServerName)
	suite.Assert().Equal("TLS1.1", suite.configuration.SMTP.TLS.MinimumVersion)
	suite.Assert().False(suite.configuration.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldRaiseErrOnInvalidSender() {
	suite.configuration.SMTP.Sender = "Google <test@example.com>"

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "smtp notifier: the sender must be only an email address but is configured to 'Google <test@example.com>', if you want to configure the name of the sender please use the new sender_name option")
}

func (suite *NotifierSuite) TestSMTPShouldSetSenderToUserWhenEmailAndSenderBlank() {
	suite.configuration.SMTP.Username = "john@example.com"
	suite.configuration.SMTP.Sender = ""

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(suite.configuration.SMTP.Username, suite.configuration.SMTP.Sender)
}

func (suite *NotifierSuite) TestSMTPShouldRaiseErrOnBlankSender() {
	suite.configuration.SMTP.Sender = ""

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "smtp notifier: the 'sender' must be configured")
}

func (suite *NotifierSuite) TestSMTPShouldEnsureHostAndPortAreProvided() {
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

	suite.Assert().EqualError(errors[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "host"))
	suite.Assert().EqualError(errors[1], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "port"))
}

func (suite *NotifierSuite) TestSMTPShouldEnsureSenderIsProvided() {
	suite.configuration.FileSystem = nil

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.configuration.SMTP.Sender = ""

	ValidateNotifier(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "sender"))
}

/*
	File Tests.
*/
func (suite *NotifierSuite) TestFileShouldEnsureFilenameIsProvided() {
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

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierFileSystemFileNameNotConfigured)
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}
