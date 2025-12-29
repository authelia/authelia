package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewAuthenticationFile directly instantiates a new authentication.UserProvider using a *authentication.FileUserProvider.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewAuthenticationFile(config *schema.Configuration, filters []configuration.BytesFilter) authentication.UserProvider {
	return authentication.NewFileUserProvider(config.AuthenticationBackend.File, filters)
}

// NewAuthenticationLDAP directly instantiates a new authentication.UserProvider using a *authentication.LDAPUserProvider.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewAuthenticationLDAP(config *schema.Configuration, caCertPool *x509.CertPool) authentication.UserProvider {
	return authentication.NewLDAPUserProvider(config.AuthenticationBackend, caCertPool)
}
