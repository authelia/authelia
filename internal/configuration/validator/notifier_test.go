package validator

import (
	"fmt"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type NotifierSuite struct {
	suite.Suite
	config    schema.NotifierConfiguration
	validator *schema.StructValidator
}

func (suite *NotifierSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config.SMTP = &schema.SMTPNotifierConfiguration{
		Username: "john",
		Password: "password",
		Sender:   mail.Address{Name: "Authelia", Address: "authelia@example.com"},
		Host:     "example.com",
		Port:     25,
	}
	suite.config.FileSystem = nil
}

/*
	Common Tests.
*/
func (suite *NotifierSuite) TestShouldEnsureAtLeastSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.config.SMTP = nil

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierNotConfigured)
}

func (suite *NotifierSuite) TestShouldEnsureEitherSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasErrors())

	suite.config.FileSystem = &schema.FileSystemNotifierConfiguration{
		Filename: "test",
	}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierMultipleConfigured)
}

/*
	SMTP Tests.
*/
func (suite *NotifierSuite) TestSMTPShouldSetTLSDefaults() {
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("example.com", suite.config.SMTP.TLS.ServerName)
	suite.Assert().Equal("TLS1.2", suite.config.SMTP.TLS.MinimumVersion)
	suite.Assert().False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldDefaultTLSServerNameToHost() {
	suite.config.SMTP.Host = "google.com"
	suite.config.SMTP.TLS = &schema.TLSConfig{
		MinimumVersion: "TLS1.1",
	}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("google.com", suite.config.SMTP.TLS.ServerName)
	suite.Assert().Equal("TLS1.1", suite.config.SMTP.TLS.MinimumVersion)
	suite.Assert().False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldEnsureHostAndPortAreProvided() {
	suite.config.FileSystem = nil
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.config.SMTP.Host = ""
	suite.config.SMTP.Port = 0

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().True(suite.validator.HasErrors())

	errors := suite.validator.Errors()

	suite.Require().Len(errors, 2)

	suite.Assert().EqualError(errors[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "host"))
	suite.Assert().EqualError(errors[1], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "port"))
}

func (suite *NotifierSuite) TestSMTPShouldEnsureSenderIsProvided() {
	suite.config.SMTP.Sender = mail.Address{}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "sender"))
}

/*
	File Tests.
*/
func (suite *NotifierSuite) TestFileShouldEnsureFilenameIsProvided() {
	suite.config.SMTP = nil
	suite.config.FileSystem = &schema.FileSystemNotifierConfiguration{
		Filename: "test",
	}
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.config.FileSystem.Filename = ""

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierFileSystemFileNameNotConfigured)
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}
