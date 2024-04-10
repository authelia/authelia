package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func NewAuthenticationFile(config *schema.Configuration) authentication.UserProvider {
	return authentication.NewFileUserProvider(config.AuthenticationBackend.File)
}

func NewAuthenticationLDAP(config *schema.Configuration, caCertPool *x509.CertPool) authentication.UserProvider {
	return authentication.NewLDAPUserProvider(config.AuthenticationBackend, caCertPool)
}
