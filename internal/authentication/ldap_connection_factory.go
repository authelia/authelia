package authentication

import (
	"github.com/go-ldap/ldap/v3"
)

// ProductionLDAPConnectionFactory the production implementation of an ldap connection factory.
type ProductionLDAPConnectionFactory struct{}

// NewProductionLDAPConnectionFactory create a concrete ldap connection factory.
func NewProductionLDAPConnectionFactory() *ProductionLDAPConnectionFactory {
	return &ProductionLDAPConnectionFactory{}
}

// DialURL creates a connection from an LDAP URL when successful.
func (f *ProductionLDAPConnectionFactory) DialURL(addr string, opts ...ldap.DialOpt) (conn LDAPConnection, err error) {
	return ldap.DialURL(addr, opts...)
}
