package middlewares

import (
	"crypto/x509"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/notification"
	"github.com/authelia/authelia/v4/internal/ntp"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/totp"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// SetContentTypeApplicationJSON sets the Content-Type header to `application/json; charset=utf-8`.
func SetContentTypeApplicationJSON(ctx *fasthttp.RequestCtx) {
	ctx.SetContentTypeBytes(contentTypeApplicationJSON)
}

// SetContentTypeTextPlain sets the Content-Type header to `text/plain; charset=utf-8`.
func SetContentTypeTextPlain(ctx *fasthttp.RequestCtx) {
	ctx.SetContentTypeBytes(contentTypeTextPlain)
}

// NewProviders provisions all providers based on the configuration provided.
func NewProviders(config *schema.Configuration, caCertPool *x509.CertPool) (providers Providers, warns, errs []error) {
	providers.Random = &random.Cryptographical{}
	providers.StorageProvider = storage.NewProvider(config, caCertPool)
	providers.Authorizer = authorization.NewAuthorizer(config)
	providers.NTP = ntp.NewProvider(&config.NTP)
	providers.PasswordPolicy = NewPasswordPolicyProvider(config.PasswordPolicy)
	providers.Regulator = regulation.NewRegulator(config.Regulation, providers.StorageProvider, clock.New())
	providers.SessionProvider = session.NewProvider(config.Session, caCertPool, providers.StorageProvider)
	providers.TOTP = totp.NewTimeBasedProvider(config.TOTP)
	providers.UserAttributeResolver = expression.NewUserAttributes(config)
	providers.UserProvider = NewAuthenticationProvider(config, caCertPool)

	var err error
	if providers.Templates, err = templates.New(templates.Config{EmailTemplatesPath: config.Notifier.TemplatePath}); err != nil {
		errs = append(errs, err)
	}

	if providers.MetaDataService, err = webauthn.NewMetaDataProvider(config, providers.StorageProvider); err != nil {
		errs = append(errs, err)
	}

	switch {
	case config.Notifier.SMTP != nil:
		providers.Notifier = notification.NewSMTPNotifier(config.Notifier.SMTP, caCertPool)
	case config.Notifier.FileSystem != nil:
		providers.Notifier = notification.NewFileNotifier(*config.Notifier.FileSystem)
	}

	providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, providers.StorageProvider, providers.Templates)

	if config.Telemetry.Metrics.Enabled {
		providers.Metrics = metrics.NewPrometheus()
	}

	return providers, warns, errs
}

// NewAuthenticationProvider returns a new authentication.UserProvider.
func NewAuthenticationProvider(config *schema.Configuration, caCertPool *x509.CertPool) (provider authentication.UserProvider) {
	switch {
	case config.AuthenticationBackend.File != nil:
		return authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	case config.AuthenticationBackend.LDAP != nil:
		return authentication.NewLDAPUserProvider(config.AuthenticationBackend, caCertPool)
	default:
		return nil
	}
}
