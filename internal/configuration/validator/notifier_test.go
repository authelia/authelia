package validator

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type NotifierSuite struct {
	suite.Suite
	config    schema.Notifier
	validator *schema.StructValidator
}

func (suite *NotifierSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config.SMTP = &schema.NotifierSMTP{
		Address:  &schema.AddressSMTP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeSMTP, exampleDotCom, 25)},
		Username: "john",
		Password: "password",
		Sender:   mail.Address{Name: "Authelia", Address: "authelia@example.com"},
	}
	suite.config.FileSystem = nil
}

/*
Common Tests.
*/
func (suite *NotifierSuite) TestShouldEnsureAtLeastSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.config.SMTP = nil

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().True(suite.validator.HasErrors())

	suite.Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], errFmtNotifierNotConfigured)
}

func (suite *NotifierSuite) TestShouldEnsureEitherSMTPOrFilesystemIsProvided() {
	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Errors(), 0)

	suite.config.FileSystem = &schema.NotifierFileSystem{
		Filename: "test",
	}

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().True(suite.validator.HasErrors())

	suite.Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], errFmtNotifierMultipleConfigured)
}

/*
SMTP Tests.
*/
func (suite *NotifierSuite) TestSMTPShouldSetTLSDefaults() {
	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(exampleDotCom, suite.config.SMTP.TLS.ServerName)
	suite.Equal(uint16(tls.VersionTLS12), suite.config.SMTP.TLS.MinimumVersion.Value)
	suite.False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldSetDefaultsWithLegacyAddress() {
	suite.config.SMTP.Address = nil
	suite.config.SMTP.Host = "xyz" //nolint:staticcheck
	suite.config.SMTP.Port = 123   //nolint:staticcheck

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(&schema.AddressSMTP{Address: MustParseAddress("smtp://xyz:123")}, suite.config.SMTP.Address)
	suite.Equal("xyz", suite.config.SMTP.TLS.ServerName)
	suite.Equal(uint16(tls.VersionTLS12), suite.config.SMTP.TLS.MinimumVersion.Value)
	suite.False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldErrorWithAddressAndLegacyAddress() {
	suite.config.SMTP.Host = "fgh" //nolint:staticcheck
	suite.config.SMTP.Port = 123   //nolint:staticcheck

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Equal(&schema.AddressSMTP{Address: MustParseAddress("smtp://example.com:25")}, suite.config.SMTP.Address)
	suite.Equal(exampleDotCom, suite.config.SMTP.TLS.ServerName)
	suite.Equal(uint16(tls.VersionTLS12), suite.config.SMTP.TLS.MinimumVersion.Value)
	suite.False(suite.config.SMTP.TLS.SkipVerify)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "notifier: smtp: option 'host' and 'port' can't be configured at the same time as 'address'")
}

func (suite *NotifierSuite) TestSMTPShouldErrorWithInvalidAddressScheme() {
	suite.config.SMTP.Address = &schema.AddressSMTP{Address: MustParseAddress("udp://example.com:25")}

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Equal(&schema.AddressSMTP{Address: MustParseAddress("udp://example.com:25")}, suite.config.SMTP.Address)
	suite.Equal(exampleDotCom, suite.config.SMTP.TLS.ServerName)
	suite.Equal(uint16(tls.VersionTLS12), suite.config.SMTP.TLS.MinimumVersion.Value)
	suite.False(suite.config.SMTP.TLS.SkipVerify)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "notifier: smtp: option 'address' with value 'udp://example.com:25' is invalid: scheme must be one of 'smtp', 'submission', or 'submissions' but is configured as 'udp'")
}

func (suite *NotifierSuite) TestSMTPShouldDefaultStartupCheckAddress() {
	suite.Equal(mail.Address{Name: "", Address: ""}, suite.config.SMTP.StartupCheckAddress)

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(mail.Address{Name: "Authelia Test", Address: "test@authelia.com"}, suite.config.SMTP.StartupCheckAddress)
}

func (suite *NotifierSuite) TestSMTPShouldDefaultTLSServerNameToHost() {
	suite.config.SMTP.Address.SetHostname("google.com")
	suite.config.SMTP.TLS = &schema.TLS{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS11},
	}

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("google.com", suite.config.SMTP.TLS.ServerName)
	suite.Equal(uint16(tls.VersionTLS11), suite.config.SMTP.TLS.MinimumVersion.MinVersion())
	suite.False(suite.config.SMTP.TLS.SkipVerify)
}

func (suite *NotifierSuite) TestSMTPShouldErrorOnSSL30() {
	suite.config.SMTP.TLS = &schema.TLS{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
	}

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "notifier: smtp: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func (suite *NotifierSuite) TestSMTPShouldErrorOnTLSMinVerGreaterThanMaxVer() {
	suite.config.SMTP.TLS = &schema.TLS{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
		MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
	}

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "notifier: smtp: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS 1.3 is greater than the maximum version TLS 1.0")
}

func (suite *NotifierSuite) TestSMTPShouldWarnOnDisabledSTARTTLS() {
	suite.config.SMTP.DisableStartTLS = true

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Require().Len(suite.validator.Warnings(), 1)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualError(suite.validator.Warnings()[0], "notifier: smtp: option 'disable_starttls' is enabled: opportunistic STARTTLS is explicitly disabled which means all emails will be sent insecurely over plaintext and this setting is only necessary for non-compliant SMTP servers which advertise they support STARTTLS when they actually don't support STARTTLS")
}

func (suite *NotifierSuite) TestSMTPShouldEnsureHostAndPortAreProvided() {
	suite.config.FileSystem = nil
	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.config.SMTP.Address = nil

	ValidateNotifier(&suite.config, suite.validator, nil)

	errors := suite.validator.Errors()

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(errors, 1)

	suite.EqualError(errors[0], "notifier: smtp: option 'address' is required")
}

func (suite *NotifierSuite) TestSMTPShouldEnsureSenderIsProvided() {
	suite.config.SMTP.Sender = mail.Address{}

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().True(suite.validator.HasErrors())

	suite.Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], fmt.Sprintf(errFmtNotifierSMTPNotConfigured, "sender"))
}

func (suite *NotifierSuite) TestTemplatesEmptyDir() {
	dir := suite.T().TempDir()

	suite.config.TemplatePath = dir

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *NotifierSuite) TestTemplatesEmptyDirNoExist() {
	dir := suite.T().TempDir()

	p := filepath.Join(dir, "notexist")

	suite.config.TemplatePath = p

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 1)

	assert.EqualError(suite.T(), suite.validator.Errors()[0], fmt.Sprintf("notifier: option 'template_path' refers to location '%s' which does not exist", p))
}

/*
File Tests.
*/
func (suite *NotifierSuite) TestFileShouldEnsureFilenameIsProvided() {
	suite.config.SMTP = nil
	suite.config.FileSystem = &schema.NotifierFileSystem{
		Filename: "test",
	}
	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.config.FileSystem.Filename = ""

	ValidateNotifier(&suite.config, suite.validator, nil)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().True(suite.validator.HasErrors())

	suite.Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], errFmtNotifierFileSystemFileNameNotConfigured)
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}

func TestNotifierMiscMissingTemplateTests(t *testing.T) {
	config := &schema.Notifier{
		TemplatePath: string([]byte{0x0, 0x1}),
	}

	validator := schema.NewStructValidator()

	validateNotifierTemplates(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "notifier: option 'template_path' refers to location '\x00\x01' which couldn't be opened: stat \x00\x01: invalid argument")
}
