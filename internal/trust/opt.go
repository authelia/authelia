package trust

import (
	"crypto/x509"
)

// ProductionOpt describes a Production option.
type ProductionOpt func(provider *Production)

// WithCertificatePaths alters the paths this provider checks for relevant trusted certificates.
func WithCertificatePaths(paths ...string) ProductionOpt {
	return func(provider *Production) {
		provider.config.Paths = paths
	}
}

// WithSystem sets the value which controls if the system certificate pool is trusted. Default is true.
func WithSystem(system bool) ProductionOpt {
	return func(provider *Production) {
		provider.config.System = system
	}
}

// WithValidationReturnErrors sets the value which determines if invalid certificates will return an error. Default is
// true.
func WithValidationReturnErrors(errs bool) ProductionOpt {
	return func(provider *Production) {
		provider.config.ValidationReturnErrors = errs
	}
}

// WithValidateNotAfter sets the value which determines if certificates not after time value (expiration) will be
// validated. Default is true.
func WithValidateNotAfter(expired bool) ProductionOpt {
	return func(provider *Production) {
		provider.config.ValidateNotAfter = expired
	}
}

// WithValidateNotBefore sets the value which determines if the certificate not before time value will be validated.
// Default is true.
func WithValidateNotBefore(future bool) ProductionOpt {
	return func(provider *Production) {
		provider.config.ValidateNotBefore = future
	}
}

// WithStatic includes static trusted certificates.
func WithStatic(static ...*x509.Certificate) ProductionOpt {
	return func(provider *Production) {
		provider.config.Static = append(provider.config.Static, static...)
	}
}
