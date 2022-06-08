package authentication

import (
	"github.com/go-ldap/ldap/v3"
)

// ProductionLDAPClientFactory the production implementation of an ldap connection factory.
type ProductionLDAPClientFactory struct{}

// NewProductionLDAPClientFactory create a concrete ldap connection factory.
func NewProductionLDAPClientFactory() *ProductionLDAPClientFactory {
	return &ProductionLDAPClientFactory{}
}

// DialURL creates a client from an LDAP URL when successful.
func (f *ProductionLDAPClientFactory) DialURL(addr string, opts ...ldap.DialOpt) (client LDAPClient, err error) {
	return ldap.DialURL(addr, opts...)
}
