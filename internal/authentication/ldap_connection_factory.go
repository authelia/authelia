package authentication

import (
	"github.com/go-ldap/ldap/v3"
)

// LDAPConnectionFactory an interface of factory of ldap connections.
type LDAPConnectionFactory interface {
	DialURL(addr string, opts ...ldap.DialOpt) (LDAPConnection, error)
}

// LDAPConnectionFactoryImpl the production implementation of an ldap connection factory.
type LDAPConnectionFactoryImpl struct{}

// NewLDAPConnectionFactoryImpl create a concrete ldap connection factory.
func NewLDAPConnectionFactoryImpl() *LDAPConnectionFactoryImpl {
	return &LDAPConnectionFactoryImpl{}
}

// DialURL creates a connection from an LDAP URL when successful.
func (lcf *LDAPConnectionFactoryImpl) DialURL(addr string, opts ...ldap.DialOpt) (LDAPConnection, error) {
	conn, err := ldap.DialURL(addr, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
