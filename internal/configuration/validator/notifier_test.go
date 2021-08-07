package validator

import (
	"fmt"
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

	suite.Assert().EqualError(suite.validator.Errors()[0], errFmtNotifierFileSystemFileNameNotConfigured)
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

	suite.Assert().EqualError(errors[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "host"))
	suite.Assert().EqualError(errors[1], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "port"))
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

	suite.Assert().EqualError(suite.validator.Errors()[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "sender"))
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}

func TestShouldEnsureFilesystemNotifierIsNotConfiguredWithSMTPSecret(t *testing.T) {
	configuration := &schema.NotifierConfiguration{
		FileSystem: &schema.FileSystemNotifierConfiguration{
			Filename: "/tmp/file.txt",
		},
		SMTP: &schema.SMTPNotifierConfiguration{
			Password: "apple",
		},
	}

	validator := schema.NewStructValidator()

	ValidateNotifier(configuration, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], errFmtNotifierMultipleConfiguredSecret)
}
