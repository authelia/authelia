package notification

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	th "html/template"
	"net/mail"
	"testing"
	tt "text/template"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestNewSMTPNotifier(t *testing.T) {
	testCases := []struct {
		name     string
		config   *schema.NotifierSMTP
		certPool *x509.CertPool
		validate func(t *testing.T, notifier *SMTPNotifier)
	}{
		{
			"ShouldHandleNormalConfig",
			&schema.NotifierSMTP{
				Address: schema.NewSMTPAddress("submission", "example.com", 587),
			},
			nil,
			nil,
		},
		{
			"ShouldHandleNormalConfigDisableStartTlS",
			&schema.NotifierSMTP{
				Address:         schema.NewSMTPAddress("submission", "example.com", 587),
				DisableStartTLS: true,
			},
			nil,
			nil,
		},
		{
			name: "ShouldHandleNormalConfigDisableRequireTlS",
			config: &schema.NotifierSMTP{
				Address:           schema.NewSMTPAddress("submission", "example.com", 587),
				DisableRequireTLS: true,
				Sender:            mail.Address{Name: "Example Name", Address: "example@example.com"},
			},
		},
		{
			"ShouldRequireExplicitTLS",
			&schema.NotifierSMTP{
				Address: schema.NewSMTPAddress("submissions", "example.com", 465),
				TLS: &schema.TLS{
					MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
					MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
					SkipVerify:     false,
					ServerName:     "example.com",
				},
			},
			nil,
			func(t *testing.T, notifier *SMTPNotifier) {
				require.NotNil(t, notifier.tls)
				assert.Equal(t, uint16(tls.VersionTLS12), notifier.tls.MinVersion)
				assert.Equal(t, uint16(tls.VersionTLS13), notifier.tls.MaxVersion)
				assert.False(t, notifier.tls.InsecureSkipVerify)
				assert.Equal(t, "example.com", notifier.tls.ServerName)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier := NewSMTPNotifier(tc.config, tc.certPool)
			require.NotNil(t, notifier)

			if tc.validate != nil {
				tc.validate(t, notifier)
			}
		})
	}
}

func SetupMockTest(t *testing.T) (ctrl *gomock.Controller, factory *MockSMTPClientFactory, client *MockSMTPClient, et *templates.EmailTemplate) {
	t.Helper()

	var err error

	et = &templates.EmailTemplate{}

	et.HTML, err = th.New("example").Parse("Example")
	require.NoError(t, err)

	et.Text, err = tt.New("example").Parse("Hello World")
	require.NoError(t, err)

	ctrl = gomock.NewController(t)

	factory = NewMockSMTPClientFactory(ctrl)

	client = NewMockSMTPClient(ctrl)

	return
}

func SetupNotifier(factory SMTPClientFactory, log *logrus.Logger) (notifier *SMTPNotifier) {
	return &SMTPNotifier{
		config: &schema.NotifierSMTP{
			Address: schema.NewSMTPAddress("submission", "example.com", 587),
		},
		random:  random.NewMathematical(),
		factory: factory,
		log:     log.WithField("notifier", "smtp"),
	}
}

func TestSMTPNotifier_StartupCheck_AllGreen(t *testing.T) {
	ctrl, factory, client, _ := SetupMockTest(t)

	defer ctrl.Finish()

	logger, hook := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(nil),
		client.EXPECT().Close().Return(nil),
	)

	assert.NoError(t, notifier.StartupCheck())

	entry := hook.LastEntry()

	require.NotNil(t, entry)
	assert.Equal(t, logrus.TraceLevel, entry.Level)
	assert.Equal(t, "Closing Startup Check Connection", entry.Message)
}

func TestSMTPNotifier_StartupCheck_ErrorClose(t *testing.T) {
	ctrl, factory, client, _ := SetupMockTest(t)

	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(nil),
		client.EXPECT().Close().Return(fmt.Errorf("bad connection")),
	)

	assert.EqualError(t, notifier.StartupCheck(), "failed to close connection: bad connection")
}

func TestSMTPNotifier_StartupCheck_ErrorDial(t *testing.T) {
	ctrl, factory, client, _ := SetupMockTest(t)

	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(fmt.Errorf("failed to dial the dude")),
	)

	assert.EqualError(t, notifier.StartupCheck(), "failed to dial connection: failed to dial the dude")
}

func TestSMTPNotifier_StartupCheck_ErrorGetClientFromFactory(t *testing.T) {
	ctrl, factory, _, _ := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(nil, fmt.Errorf("no client found")),
	)

	assert.EqualError(t, notifier.StartupCheck(), "notifier: smtp: failed to establish client: no client found")
}

func TestSMTPNotifier_ErrorSender(t *testing.T) {
	ctrl, factory, _, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	assert.EqualError(t, notifier.Send(context.Background(), mail.Address{Name: "Example One", Address: "admin@example.cm"}, "example subject", et, nil), "notifier: smtp: failed to create envelope: failed to set from address: failed to parse mail address \"<@>\": mail: invalid string")
}

func TestSMTPNotifier_Send_Success(t *testing.T) {
	ctrl, factory, client, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	notifier.config.Sender = mail.Address{
		Name:    "Admin",
		Address: "admin@example.com",
	}

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(nil),
		client.EXPECT().Send(gomock.Any()).Return(nil),
		client.EXPECT().Close().Return(nil),
	)

	assert.NoError(t, notifier.Send(context.Background(), mail.Address{Name: "Example One", Address: "admin@example.cm"}, "example subject", et, nil))
}

func TestSMTPNotifier_Send_SuccessDisableHTML(t *testing.T) {
	ctrl, factory, client, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	notifier.config.Sender = mail.Address{
		Name:    "Admin",
		Address: "admin@example.com",
	}

	notifier.config.DisableHTMLEmails = true

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(nil),
		client.EXPECT().Send(gomock.Any()).Return(nil),
		client.EXPECT().Close().Return(nil),
	)

	assert.NoError(t, notifier.Send(context.Background(), mail.Address{Name: "Example One", Address: "admin@example.cm"}, "example subject", et, nil))
}

func TestSMTPNotifier_Send_ErrorRecipient(t *testing.T) {
	ctrl, factory, _, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	notifier.config.Sender = mail.Address{
		Name:    "Admin",
		Address: "admin@example.com",
	}

	assert.EqualError(t, notifier.Send(context.Background(), mail.Address{}, "example subject", et, nil), "notifier: smtp: failed to create envelope: failed to set to address: failed to parse mail address \"<@>\": mail: invalid string")
}

func TestSMTPNotifier_Send_ErrorClose(t *testing.T) {
	ctrl, factory, client, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	notifier.config.Sender = mail.Address{
		Name:    "Admin",
		Address: "admin@example.com",
	}

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(nil),
		client.EXPECT().Send(gomock.Any()).Return(nil),
		client.EXPECT().Close().Return(fmt.Errorf("bad connection")),
	)

	assert.EqualError(t, notifier.Send(context.Background(), mail.Address{Name: "Example One", Address: "admin@example.cm"}, "example subject", et, nil), "notifier: smtp: failed to close connection: bad connection")
}

func TestSMTPNotifier_Send_ErrorSend(t *testing.T) {
	ctrl, factory, client, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	notifier.config.Sender = mail.Address{
		Name:    "Admin",
		Address: "admin@example.com",
	}

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(nil),
		client.EXPECT().Send(gomock.Any()).Return(fmt.Errorf("no way mister")),
	)

	assert.EqualError(t, notifier.Send(context.Background(), mail.Address{Name: "Example One", Address: "admin@example.cm"}, "example subject", et, nil), "notifier: smtp: failed to send message: no way mister")
}

func TestSMTPNotifier_Send_ErrorDial(t *testing.T) {
	ctrl, factory, client, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	notifier.config.Sender = mail.Address{
		Name:    "Admin",
		Address: "admin@example.com",
	}

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(client, nil),
		client.EXPECT().DialWithContext(context.Background()).Return(fmt.Errorf("DIND NX container: there's no way it's DIND")),
	)

	assert.EqualError(t, notifier.Send(context.Background(), mail.Address{Name: "Example One", Address: "admin@example.cm"}, "example subject", et, nil), "notifier: smtp: failed to dial connection: DIND NX container: there's no way it's DIND")
}

func TestSMTPNotifier_Send_ErrorGetClient(t *testing.T) {
	ctrl, factory, _, et := SetupMockTest(t)
	defer ctrl.Finish()

	logger, _ := test.NewNullLogger()
	logger.Level = logrus.TraceLevel

	notifier := SetupNotifier(factory, logger)

	notifier.config.Sender = mail.Address{
		Name:    "Admin",
		Address: "admin@example.com",
	}

	gomock.InOrder(
		factory.EXPECT().GetClient().Return(nil, fmt.Errorf("no client for you, 1 year")),
	)

	assert.EqualError(t, notifier.Send(context.Background(), mail.Address{Name: "Example One", Address: "admin@example.cm"}, "example subject", et, nil), "notifier: smtp: failed to establish client: no client for you, 1 year")
}
