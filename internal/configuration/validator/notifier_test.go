// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"crypto/tls"
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
		Host:     exampleDotCom,
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

	suite.Assert().Equal(exampleDotCom, suite.config.SMTP.TLS.ServerName)
	suite.Assert().Equal(uint16(tls.VersionTLS12), suite.config.SMTP.TLS.MinimumVersion.Value)
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
		MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS11},
	}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("google.com", suite.config.SMTP.TLS.ServerName)
	suite.Assert().Equal(uint16(tls.VersionTLS11), suite.config.SMTP.TLS.MinimumVersion.MinVersion())
	suite.Assert().False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldErrorOnSSL30() {
	suite.config.SMTP.Host = exampleDotCom
	suite.config.SMTP.TLS = &schema.TLSConfig{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
	}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "notifier: smtp: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func (suite *NotifierSuite) TestSMTPShouldErrorOnTLSMinVerGreaterThanMaxVer() {
	suite.config.SMTP.Host = exampleDotCom
	suite.config.SMTP.TLS = &schema.TLSConfig{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
		MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
	}

	ValidateNotifier(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "notifier: smtp: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS1.3 is greater than the maximum version TLS1.0")
}

func (suite *NotifierSuite) TestSMTPShouldWarnOnDisabledSTARTTLS() {
	suite.config.SMTP.Host = exampleDotCom
	suite.config.SMTP.DisableStartTLS = true

	ValidateNotifier(&suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 1)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().EqualError(suite.validator.Warnings()[0], "notifier: smtp: option 'disable_starttls' is enabled: opportunistic STARTTLS is explicitly disabled which means all emails will be sent insecurely over plaintext and this setting is only necessary for non-compliant SMTP servers which advertise they support STARTTLS when they actually don't support STARTTLS")
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
