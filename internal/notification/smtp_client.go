package notification

import (
	"context"
	"crypto/tls"

	"github.com/wneessen/go-mail"
	"github.com/wneessen/go-mail/smtp"
)

type SMTPClientFactory interface {
	GetClient() (client SMTPClient, err error)
}

type SMTPClient interface {
	TLSPolicy() (policy string)
	ServerAddr() (addr string)
	SetTLSPolicy(policy mail.TLSPolicy)
	SetTLSPortPolicy(policy mail.TLSPolicy)
	SetSSL(ssl bool)
	SetSSLPort(ssl bool, fallback bool)
	SetDebugLog(val bool)
	SetTLSConfig(tlsconfig *tls.Config) (err error)
	SetUsername(username string)
	SetPassword(password string)
	SetSMTPAuth(authType mail.SMTPAuthType)
	SetSMTPAuthCustom(smtpAuth smtp.Auth)
	SetLogAuthData(logAuth bool)
	DialWithContext(ctxDial context.Context) (err error)
	Close() (err error)
	Reset() (err error)
	DialAndSend(messages ...*mail.Msg) (err error)
	DialAndSendWithContext(ctx context.Context, messages ...*mail.Msg) (err error)
	Send(messages ...*mail.Msg) (err error)
}
