package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type NotifierSuite struct {
	suite.Suite

	configuration schema.NotifierConfiguration
}

func (s *NotifierSuite) SetupTest() {
	s.configuration.SMTP = &schema.SMTPNotifierConfiguration{
		Username: "john",
		Password: "password",
		Sender:   "admin@example.com",
		Host:     "example.com",
		Port:     25,
	}
}

func (s *NotifierSuite) TestShouldEnsureAtLeastSMTPOrFilesystemIsProvided() {
	validator := schema.NewStructValidator()
	ValidateNotifier(&s.configuration, validator)

	errors := validator.Errors()
	s.Require().Len(errors, 0)

	s.configuration.SMTP = nil

	ValidateNotifier(&s.configuration, validator)

	errors = validator.Errors()
	s.Require().Len(errors, 1)
	s.Assert().EqualError(errors[0], "Notifier should be either `smtp` or `filesystem`")
}

func (s *NotifierSuite) TestShouldEnsureEitherSMTPOrFilesystemIsProvided() {
	validator := schema.NewStructValidator()
	ValidateNotifier(&s.configuration, validator)

	errors := validator.Errors()
	s.Require().Len(errors, 0)

	s.configuration.FileSystem = &schema.FileSystemNotifierConfiguration{
		Filename: "test",
	}

	ValidateNotifier(&s.configuration, validator)

	errors = validator.Errors()
	s.Require().Len(errors, 1)
	s.Assert().EqualError(errors[0], "Notifier should be either `smtp` or `filesystem`")
}

func (s *NotifierSuite) TestShouldEnsureFilenameOfFilesystemNotifierIsProvided() {
	validator := schema.NewStructValidator()

	s.configuration.SMTP = nil
	s.configuration.FileSystem = &schema.FileSystemNotifierConfiguration{
		Filename: "test",
	}
	ValidateNotifier(&s.configuration, validator)

	errors := validator.Errors()
	s.Require().Len(errors, 0)

	s.configuration.FileSystem.Filename = ""

	ValidateNotifier(&s.configuration, validator)

	errors = validator.Errors()
	s.Require().Len(errors, 1)
	s.Assert().EqualError(errors[0], "Filename of filesystem notifier must not be empty")
}

func (s *NotifierSuite) TestShouldEnsureHostAndPortOfSMTPNotifierAreProvided() {
	s.configuration.FileSystem = nil
	validator := schema.NewStructValidator()
	ValidateNotifier(&s.configuration, validator)

	errors := validator.Errors()
	s.Require().Len(errors, 0)

	s.configuration.SMTP.Host = ""
	s.configuration.SMTP.Port = 0

	ValidateNotifier(&s.configuration, validator)

	errors = validator.Errors()
	s.Require().Len(errors, 2)
	s.Assert().EqualError(errors[0], "Host of SMTP notifier must be provided")
	s.Assert().EqualError(errors[1], "Port of SMTP notifier must be provided")
}

func (s *NotifierSuite) TestShouldEnsureSenderOfSMTPNotifierAreProvided() {
	s.configuration.FileSystem = nil

	validator := schema.NewStructValidator()
	ValidateNotifier(&s.configuration, validator)

	errors := validator.Errors()
	s.Require().Len(errors, 0)

	s.configuration.SMTP.Sender = ""

	ValidateNotifier(&s.configuration, validator)

	errors = validator.Errors()
	s.Require().Len(errors, 1)
	s.Assert().EqualError(errors[0], "Sender of SMTP notifier must be provided")
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}
