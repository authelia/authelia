package notification

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/mail"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	gomail "github.com/wneessen/go-mail"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewSMTPNotifier creates a SMTPNotifier using the notifier configuration.
func NewSMTPNotifier(config *schema.NotifierSMTP, certPool *x509.CertPool) *SMTPNotifier {
	log := logging.Logger().WithFields(map[string]any{"provider": "notifier"})

	var configTLS *tls.Config

	if config.TLS != nil {
		configTLS = utils.NewTLSConfig(config.TLS, certPool)
	}

	var opts []gomail.Option

	switch {
	case config.Address.IsExplicitlySecure():
		opts = []gomail.Option{
			gomail.WithSSLPort(false),
			gomail.WithTLSPortPolicy(gomail.TLSMandatory),
		}

		log.Trace("Configuring with Explicit TLS")
	case config.DisableStartTLS:
		opts = []gomail.Option{
			gomail.WithTLSPortPolicy(gomail.NoTLS),
			gomail.WithPort(int(config.Address.Port())),
		}

		log.Trace("Configuring without TLS")
	case config.DisableRequireTLS:
		opts = []gomail.Option{
			gomail.WithTLSPortPolicy(gomail.TLSOpportunistic),
		}

		log.Trace("Configuring with Opportunistic TLS")
	default:
		opts = []gomail.Option{
			gomail.WithTLSPortPolicy(gomail.TLSMandatory),
		}

		log.Trace("Configuring with Mandatory TLS")
	}

	opts = append(opts,
		gomail.WithTLSConfig(configTLS),
		gomail.WithTimeout(config.Timeout),
		gomail.WithHELO(config.Identifier),
		gomail.WithoutNoop(),
		gomail.WithPort(int(config.Address.Port())),
	)

	var domain string

	at := strings.Index(config.Sender.Address, "@")

	if at >= 0 {
		domain = config.Sender.Address[at+1:]
	} else {
		domain = "localhost.localdomain"
	}

	log.WithFields(map[string]any{
		"port":    config.Address.Port(),
		"helo":    config.Identifier,
		"timeout": config.Timeout.Seconds(),
		"domain":  domain,
	}).Trace("Configuring Provider")

	return &SMTPNotifier{
		factory: &StandardSMTPClientFactory{
			config: config,
			opts:   opts,
		},
		config: config,
		domain: domain,
		random: random.New(),
		tls:    configTLS,
		log:    log,
		opts:   opts,
	}
}

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	factory SMTPClientFactory
	config  *schema.NotifierSMTP
	domain  string
	random  random.Provider
	tls     *tls.Config
	log     *logrus.Entry
	opts    []gomail.Option
}

// StartupCheck implements model.StartupCheck to perform startup check operations.
func (n *SMTPNotifier) StartupCheck() (err error) {
	n.log.WithFields(map[string]any{"hostname": n.config.Address.Hostname()}).Trace("Creating Startup Check Client")

	var client SMTPClient

	if client, err = n.factory.GetClient(); err != nil {
		return fmt.Errorf("notifier: smtp: failed to establish client: %w", err)
	}

	n.log.Trace("Dialing Startup Check Connection")

	if err = client.DialWithContext(context.Background()); err != nil {
		return fmt.Errorf("failed to dial connection: %w", err)
	}

	n.log.Trace("Closing Startup Check Connection")

	if err = client.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

// Send a notification via the SMTPNotifier.
func (n *SMTPNotifier) Send(ctx context.Context, recipient mail.Address, subject string, et *templates.EmailTemplate, data any) (err error) {
	var (
		msg    *gomail.Msg
		client SMTPClient
	)

	if msg, err = n.msg(recipient, subject, et, data); err != nil {
		return fmt.Errorf("notifier: smtp: failed to create envelope: %w", err)
	}

	if client, err = n.factory.GetClient(); err != nil {
		return fmt.Errorf("notifier: smtp: failed to establish client: %w", err)
	}

	if err = client.DialWithContext(ctx); err != nil {
		return fmt.Errorf("notifier: smtp: failed to dial connection: %w", err)
	}

	if err = client.Send(msg); err != nil {
		return fmt.Errorf("notifier: smtp: failed to send message: %w", err)
	}

	if err = client.Close(); err != nil {
		return fmt.Errorf("notifier: smtp: failed to close connection: %w", err)
	}

	return nil
}

func (n *SMTPNotifier) msg(recipient mail.Address, subject string, et *templates.EmailTemplate, data any) (msg *gomail.Msg, err error) {
	msg = gomail.NewMsg(
		gomail.WithMIMEVersion(gomail.MIME10),
		gomail.WithBoundary(n.random.StringCustom(30, random.CharSetAlphaNumeric)),
	)

	n.setMessageID(msg)

	if err = msg.From(n.config.Sender.String()); err != nil {
		return nil, fmt.Errorf("failed to set from address: %w", err)
	}

	if err = msg.AddTo(recipient.String()); err != nil {
		return nil, fmt.Errorf("failed to set to address: %w", err)
	}

	msg.Subject(strings.ReplaceAll(n.config.Subject, "{title}", subject))

	switch {
	case n.config.DisableHTMLEmails:
		if err = msg.SetBodyTextTemplate(et.Text, data); err != nil {
			return nil, fmt.Errorf("failed to set body: text template errored: %w", err)
		}
	default:
		if err = msg.AddAlternativeTextTemplate(et.Text, data); err != nil {
			return nil, fmt.Errorf("failed to set body: text template errored: %w", err)
		}

		if err = msg.AddAlternativeHTMLTemplate(et.HTML, data); err != nil {
			return nil, fmt.Errorf("failed to set body: html template errored: %w", err)
		}
	}

	return msg, nil
}

func (n *SMTPNotifier) setMessageID(msg *gomail.Msg) {
	rn := n.random.Intn(100000000)
	rm := n.random.Intn(10000)
	rs := n.random.StringCustom(17, random.CharSetAlphaNumeric)
	pid := os.Getpid() + rm

	msg.SetMessageIDWithValue(fmt.Sprintf("%d.%d%d.%s@%s", pid, rn, rm, rs, n.domain))
}

type StandardSMTPClientFactory struct {
	config *schema.NotifierSMTP
	opts   []gomail.Option
}

func (f *StandardSMTPClientFactory) GetClient() (client SMTPClient, err error) {
	if client, err = gomail.NewClient(f.config.Address.Hostname(), f.opts...); err != nil {
		return nil, err
	}

	switch {
	case len(f.config.Username)+len(f.config.Password) > 0:
		client.SetSMTPAuthCustom(NewOpportunisticSMTPAuth(f.config))
	default:
		client.SetSMTPAuth(gomail.SMTPAuthNoAuth)
	}

	return client, nil
}
