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
				MinimumVersion: MustParseTLSVersion("TLS1.1"),
			},
		},
	}

	notifier := NewSMTPNotifier(config.SMTP, nil, &templates.Provider{})

	assert.Equal(t, "smtp.example.com", notifier.configTLS.ServerName)
	assert.Equal(t, uint16(tls.VersionTLS11), notifier.configTLS.MinVersion)
	assert.False(t, notifier.configTLS.InsecureSkipVerify)
}

func TestShouldConfigureSMTPNotifierWithServerNameOverride(t *testing.T) {
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

	assert.Equal(t, "smtp.golang.org", notifier.configTLS.ServerName)
	assert.False(t, notifier.configTLS.InsecureSkipVerify)
}

func MustParseTLSVersion(value string) schema.TLSVersion {
	v, err := schema.NewTLSVersion(value)
	if err != nil {
		panic(err)
	}

	return *v
}
