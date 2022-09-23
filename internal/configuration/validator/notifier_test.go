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

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.config.SMTP = nil

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierNotConfigured)
}

func (suite *NotifierSuite) TestShouldEnsureEitherSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.config.FileSystem = &schema.FileSystemNotifierConfiguration{
		Filename: "test",
	}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierMultipleConfigured)
}

/*
SMTP Tests.
*/
func (suite *NotifierSuite) TestSMTPShouldSetTLSDefaults() {
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("example.com", suite.config.SMTP.TLS.ServerName)
	suite.Assert().Equal(MustParseTLSVersion("TLS1.2"), suite.config.SMTP.TLS.MinimumVersion)
	suite.Assert().False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldDefaultStartupCheckAddress() {
	suite.Assert().Equal(mail.Address{Name: "", Address: ""}, suite.config.SMTP.StartupCheckAddress)

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(mail.Address{Name: "Authelia Test", Address: "test@authelia.com"}, suite.config.SMTP.StartupCheckAddress)
}

func (suite *NotifierSuite) TestSMTPShouldDefaultTLSServerNameToHost() {
	suite.config.SMTP.Host = "google.com"
	suite.config.SMTP.TLS = &schema.TLSConfig{
		MinimumVersion: MustParseTLSVersion("TLS1.1"),
	}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("google.com", suite.config.SMTP.TLS.ServerName)
	suite.Assert().Equal(MustParseTLSVersion("TLS1.1"), suite.config.SMTP.TLS.MinimumVersion)
	suite.Assert().False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldEnsureHostAndPortAreProvided() {
	suite.config.FileSystem = nil
	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.config.SMTP.Host = ""
	suite.config.SMTP.Port = 0

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().True(suite.validator.HasErrors())

	errors := suite.validator.Errors()

	suite.Require().Len(errors, 2)

	suite.Assert().EqualError(errors[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "host"))
	suite.Assert().EqualError(errors[1], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "port"))
}

func (suite *NotifierSuite) TestSMTPShouldEnsureSenderIsProvided() {
	suite.config.SMTP.Sender = mail.Address{}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
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

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.config.FileSystem.Filename = ""

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().True(suite.validator.HasErrors())

	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierFileSystemFileNameNotConfigured)
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}

func MustParseTLSVersion(value string) schema.TLSVersion {
	v, err := schema.NewTLSVersion(value)
	if err != nil {
		panic(err)
	}

	return *v
}
