package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
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
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// New returns a completely new set of providers using the internal API. It is expected you'll check the errs return
// value for any errors, and handle any warnings in a graceful way. If errors are returned the providers should not be
// utilized to run anything.
func New(config *schema.Configuration, caCertPool *x509.CertPool) (providers middlewares.Providers, warns []error, errs []error) {
	return middlewares.NewProviders(config, caCertPool)
}

// NewClock creates a new clock provider.
func NewClock() clock.Provider {
	return clock.New()
}

// NewAuthorizer creates a new *authorization.Authorizer.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewAuthorizer(config *schema.Configuration) *authorization.Authorizer {
	return authorization.NewAuthorizer(config)
}

// NewSession creates a new *session.Provider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewSession(config *schema.Configuration, caCertPool *x509.CertPool, storageProvider storage.Provider) *session.Provider {
	return session.NewProvider(config.Session, caCertPool, storageProvider)
}

// NewRegulator creates a new *regulation.Regulator given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewRegulator(config *schema.Configuration, storage storage.RegulatorProvider, clock clock.Provider) *regulation.Regulator {
	return regulation.NewRegulator(config.Regulation, storage, clock)
}

// NewMetrics creates a new metrics.Provider.
func NewMetrics() metrics.Provider {
	return metrics.NewPrometheus()
}

// NewNTP creates a new *ntp.Provider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewNTP(config *schema.Configuration) *ntp.Provider {
	return ntp.NewProvider(&config.NTP)
}

// NewOpenIDConnect creates a new *oidc.OpenIDConnectProvider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewOpenIDConnect(config *schema.Configuration, storage storage.Provider, templates *templates.Provider) *oidc.OpenIDConnectProvider {
	return oidc.NewOpenIDConnectProvider(config, storage, templates)
}

// NewTemplates creates a new *templates.Provider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewTemplates(config *schema.Configuration) (provider *templates.Provider, err error) {
	return templates.New(templates.Config{EmailTemplatesPath: config.Notifier.TemplatePath})
}

// NewTOTP creates a new totp.Provider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewTOTP(config *schema.Configuration) totp.Provider {
	return totp.NewTimeBasedProvider(config.TOTP)
}

// NewPasswordPolicy creates a new middlewares.PasswordPolicyProvider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewPasswordPolicy(config *schema.Configuration) middlewares.PasswordPolicyProvider {
	return middlewares.NewPasswordPolicyProvider(config.PasswordPolicy)
}

// NewRandom creates a new random.Provider given a valid configuration. This uses the rand/crypto package.
func NewRandom() random.Provider {
	return &random.Cryptographical{}
}

// NewUserAttributeResolver creates a new expression.UserAttributeResolver given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewUserAttributeResolver(config *schema.Configuration) expression.UserAttributeResolver {
	return expression.NewUserAttributes(config)
}

// NewMetaDataService creates a new webauthn.MetaDataProvider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewMetaDataService(config *schema.Configuration, store storage.CachedDataProvider) (provider webauthn.MetaDataProvider, err error) {
	return webauthn.NewMetaDataProvider(config, store)
}
