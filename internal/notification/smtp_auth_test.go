package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wneessen/go-mail"
	"github.com/wneessen/go-mail/smtp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewOpportunisticSMTPAuth(t *testing.T) {
	testCases := []struct {
		name       string
		config     *schema.NotifierSMTP
		preference []mail.SMTPAuthType
		expected   *OpportunisticSMTPAuth
	}{
		{
			"ShouldHandleNormal",
			&schema.NotifierSMTP{
				Address:  schema.NewSMTPAddress("submission", "example.com", 587),
				Username: "admin",
				Password: "admin",
			},
			nil,
			&OpportunisticSMTPAuth{
				username: "admin",
				password: "admin",
				host:     "example.com",
			},
		},
		{
			"ShouldHandleNormalWithPreference",
			&schema.NotifierSMTP{
				Address:  schema.NewSMTPAddress("submission", "example.com", 587),
				Username: "admin",
				Password: "admin",
			},
			[]mail.SMTPAuthType{mail.SMTPAuthCramMD5},
			&OpportunisticSMTPAuth{
				username:      "admin",
				password:      "admin",
				host:          "example.com",
				satPreference: []mail.SMTPAuthType{mail.SMTPAuthCramMD5},
			},
		},
		{
			"ShouldHandleReturnNil",
			&schema.NotifierSMTP{
				Address:  schema.NewSMTPAddress("submission", "example.com", 587),
				Username: "",
				Password: "",
			},
			nil,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewOpportunisticSMTPAuth(tc.config, tc.preference...)

			if tc.expected != nil {
				require.NotNil(t, actual)

				oactual, ok := actual.(*OpportunisticSMTPAuth)
				require.True(t, ok)

				assert.Equal(t, tc.expected.username, oactual.username)
				assert.Equal(t, tc.expected.password, oactual.password)
				assert.Equal(t, tc.expected.host, oactual.host)
				assert.Equal(t, tc.expected.disableRequireTLS, oactual.disableRequireTLS)
				assert.Equal(t, tc.expected.satPreference, oactual.satPreference)
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}

func TestOpportunisticSMTPAuth_SetPreferred(t *testing.T) {
	testCases := []struct {
		name       string
		preference []mail.SMTPAuthType
		have       *smtp.ServerInfo
		expected   smtp.Auth
	}{
		{
			"ShouldNotSetAnything",
			nil,
			&smtp.ServerInfo{
				Auth: []string{"PLAIN"},
			},
			nil,
		},
		{
			"ShouldSetPlainAuth",
			[]mail.SMTPAuthType{mail.SMTPAuthPlain},
			&smtp.ServerInfo{
				Auth: []string{"PLAIN", "SCRAM-SHA-256"},
			},
			smtp.PlainAuth("", "admin", "password", "example.com", false),
		},
		{
			"ShouldSetLoginAuth",
			[]mail.SMTPAuthType{mail.SMTPAuthLogin},
			&smtp.ServerInfo{
				Auth: []string{"LOGIN", "SCRAM-SHA-256"},
			},
			smtp.LoginAuth("admin", "password", "example.com", false),
		},
		{
			"ShouldSetSCRAMSHA256Auth",
			[]mail.SMTPAuthType{mail.SMTPAuthSCRAMSHA256},
			&smtp.ServerInfo{
				Auth: []string{"PLAIN", "SCRAM-SHA-256"},
			},
			smtp.ScramSHA256Auth("admin", "password"),
		},
		{
			"ShouldSetSCRAMSHA1Auth",
			[]mail.SMTPAuthType{mail.SMTPAuthSCRAMSHA1},
			&smtp.ServerInfo{
				Auth: []string{"PLAIN", "SCRAM-SHA-1"},
			},
			smtp.ScramSHA1Auth("admin", "password"),
		},
		{
			"ShouldSetCRAMMD5Auth",
			[]mail.SMTPAuthType{mail.SMTPAuthCramMD5},
			&smtp.ServerInfo{
				Auth: []string{"PLAIN", "CRAM-MD5"},
			},
			smtp.CRAMMD5Auth("admin", "password"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auth := &OpportunisticSMTPAuth{
				username:      "admin",
				password:      "password",
				host:          "example.com",
				satPreference: tc.preference,
			}

			auth.setPreferred(tc.have)

			if tc.expected == nil {
				assert.Nil(t, auth.sa)
			} else {
				assert.IsType(t, tc.expected, auth.sa)
			}
		})
	}
}

func TestOpportunisticSMTPAuth_Set(t *testing.T) {
	testCases := []struct {
		name       string
		preference []mail.SMTPAuthType
		have       *smtp.ServerInfo
		expected   smtp.Auth
	}{
		{
			"ShouldSetLoginAuth",
			nil,
			&smtp.ServerInfo{
				Auth: []string{"LOGIN"},
			},
			smtp.LoginAuth("admin", "password", "example.com", false),
		},
		{
			"ShouldSetPlainAuth",
			nil,
			&smtp.ServerInfo{
				Auth: []string{"LOGIN", "PLAIN"},
			},
			smtp.PlainAuth("", "admin", "password", "example.com", false),
		},
		{
			"ShouldSetCRAMMD5Auth",
			nil,
			&smtp.ServerInfo{
				Auth: []string{"LOGIN", "PLAIN", "CRAM-MD5"},
			},
			smtp.CRAMMD5Auth("admin", "password"),
		},
		{
			"ShouldSetSCRAMSHA1Auth",
			nil,
			&smtp.ServerInfo{
				Auth: []string{"LOGIN", "PLAIN", "CRAM-MD5", "SCRAM-SHA-1"},
			},
			smtp.ScramSHA1Auth("admin", "password"),
		},
		{
			"ShouldSetSCRAMSHA256Auth",
			nil,
			&smtp.ServerInfo{
				Auth: []string{"LOGIN", "PLAIN", "CRAM-MD5", "SCRAM-SHA-1", "SCRAM-SHA-256"},
			},
			smtp.ScramSHA256Auth("admin", "password"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auth := &OpportunisticSMTPAuth{
				username:      "admin",
				password:      "password",
				host:          "example.com",
				satPreference: tc.preference,
			}

			auth.set(tc.have)

			if tc.expected == nil {
				assert.Nil(t, auth.sa)
			} else {
				assert.IsType(t, tc.expected, auth.sa)
			}
		})
	}
}

func TestNewOpportunisticSMTPAuth_Flow(t *testing.T) {
	auth := NewOpportunisticSMTPAuth(&schema.NotifierSMTP{Username: "admin", Password: "password", Address: schema.NewSMTPAddress("submission", "example.com", 587)})

	proto, toServer, err := auth.Start(&smtp.ServerInfo{Auth: []string{"PLAIN", "SCRAM-SHA-256"}})
	require.NoError(t, err)

	assert.Equal(t, "SCRAM-SHA-256", proto)
	assert.Equal(t, []byte(nil), toServer)

	toServer, err = auth.Next([]byte("example"), false)
	require.NoError(t, err)

	assert.Equal(t, []byte(nil), toServer)
}

func TestNewOpportunisticSMTPAuth_FlowUnsupported(t *testing.T) {
	auth := NewOpportunisticSMTPAuth(&schema.NotifierSMTP{Username: "admin", Password: "password", Address: schema.NewSMTPAddress("submission", "example.com", 587)})

	proto, toServer, err := auth.Start(&smtp.ServerInfo{Auth: []string{"PLAINX", "SCRAM-XSHA-256"}})
	assert.EqualError(t, err, "unsupported SMTP AUTH types: PLAINX, SCRAM-XSHA-256")

	assert.Equal(t, "", proto)
	assert.Equal(t, []byte(nil), toServer)
}
