package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/ntp"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/totp"
)

func New(config *schema.Configuration, caCertPool *x509.CertPool) (providers middlewares.Providers, warns []error, errs []error) {
	return middlewares.NewProviders(config, caCertPool)
}

func NewClock() clock.Provider {
	return clock.New()
}

func NewAuthorizer(config *schema.Configuration) *authorization.Authorizer {
	return authorization.NewAuthorizer(config)
}

func NewSession(config *schema.Configuration, caCertPool *x509.CertPool) *session.Provider {
	return session.NewProvider(config.Session, caCertPool)
}

func NewRegulator(config *schema.Configuration, storage storage.RegulatorProvider, clock clock.Provider) *regulation.Regulator {
	return regulation.NewRegulator(config.Regulation, storage, clock)
}

func NewMetrics() metrics.Provider {
	return metrics.NewPrometheus()
}

func NewNTP(config *schema.Configuration) *ntp.Provider {
	return ntp.NewProvider(&config.NTP)
}

func NewOpenIDConnect(config *schema.IdentityProvidersOpenIDConnect, storage storage.Provider, templates *templates.Provider) *oidc.OpenIDConnectProvider {
	return oidc.NewOpenIDConnectProvider(config, storage, templates)
}

func NewTemplates(config *schema.Configuration) (provider *templates.Provider, err error) {
	return templates.New(templates.Config{EmailTemplatesPath: config.Notifier.TemplatePath})
}

func NewTOTP(config *schema.Configuration) totp.Provider {
	return totp.NewTimeBasedProvider(config.TOTP)
}

func NewPasswordPolicy(config *schema.Configuration) middlewares.PasswordPolicyProvider {
	return middlewares.NewPasswordPolicyProvider(config.PasswordPolicy)
}

func NewRandom() random.Provider {
	return &random.Cryptographical{}
}
