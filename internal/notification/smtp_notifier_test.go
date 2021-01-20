package notification

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

func TestShouldConfigureSMTPNotifierWithTLS11AndDefaultHostname(t *testing.T) {
	config := &schema.NotifierConfiguration{
		DisableStartupCheck: true,
		SMTP: &schema.SMTPNotifierConfiguration{
			Host: "smtp.example.com",
			Port: 25,
			TLS: &schema.TLSConfig{
				MinimumVersion: "TLS1.1",
			},
		},
	}

	sv := schema.NewStructValidator()
	validator.ValidateNotifier(config, sv)

	notifier := NewSMTPNotifier(*config.SMTP, nil)

	assert.Equal(t, "smtp.example.com", notifier.tlsConfig.ServerName)
	assert.Equal(t, uint16(tls.VersionTLS11), notifier.tlsConfig.MinVersion)
	assert.False(t, notifier.tlsConfig.InsecureSkipVerify)
	assert.Equal(t, "smtp.example.com:25", notifier.address)
}

func TestShouldConfigureSMTPNotifierWithServerNameOverrideAndDefaultTLS12(t *testing.T) {
	config := &schema.NotifierConfiguration{
		DisableStartupCheck: true,
		SMTP: &schema.SMTPNotifierConfiguration{
			Host: "smtp.example.com",
			Port: 25,
			TLS: &schema.TLSConfig{
				ServerName: "smtp.golang.org",
			},
		},
	}

	sv := schema.NewStructValidator()
	validator.ValidateNotifier(config, sv)

	notifier := NewSMTPNotifier(*config.SMTP, nil)

	assert.Equal(t, "smtp.golang.org", notifier.tlsConfig.ServerName)
	assert.Equal(t, uint16(tls.VersionTLS12), notifier.tlsConfig.MinVersion)
	assert.False(t, notifier.tlsConfig.InsecureSkipVerify)
	assert.Equal(t, "smtp.example.com:25", notifier.address)
}
