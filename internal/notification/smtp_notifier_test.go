package notification

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestShouldConfigureSMTPNotifierWithTLS11(t *testing.T) {
	config := &schema.NotifierConfiguration{
		DisableStartupCheck: true,
		SMTP: &schema.SMTPNotifierConfiguration{
			Host: "smtp.example.com",
			Port: 25,
			TLS: &schema.TLSConfig{
				ServerName:     "smtp.example.com",
				MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS11},
			},
		},
	}

	notifier := NewSMTPNotifier(config.SMTP, nil, &templates.Provider{})

	assert.Equal(t, "smtp.example.com", notifier.tlsConfig.ServerName)
	assert.Equal(t, uint16(tls.VersionTLS11), notifier.tlsConfig.MinVersion)
	assert.False(t, notifier.tlsConfig.InsecureSkipVerify)
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

	notifier := NewSMTPNotifier(config.SMTP, nil, &templates.Provider{})

	assert.Equal(t, "smtp.golang.org", notifier.tlsConfig.ServerName)
	assert.Equal(t, uint16(tls.VersionTLS12), notifier.tlsConfig.MinVersion)
	assert.False(t, notifier.tlsConfig.InsecureSkipVerify)
}
