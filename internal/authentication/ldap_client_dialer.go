package authentication

import "github.com/go-ldap/ldap/v3"

// LDAPClientDialer is an abstract type that dials a ldap.Client.
type LDAPClientDialer interface {
	// DialURL takes a single address and dials it returning the ldap.Client.
	DialURL(addr string, opts ...ldap.DialOpt) (client ldap.Client, err error)
}

// NewLDAPConnectionDialerStandard returns a new *LDAPClientDialerStandard.
func NewLDAPConnectionDialerStandard() *LDAPClientDialerStandard {
	return &LDAPClientDialerStandard{}
}

// LDAPClientDialerStandard is a concrete type that dials a ldap.Client and returns it, implementing the
// LDAPClientDialer.
type LDAPClientDialerStandard struct{}

// DialURL takes a single address and dials it returning the ldap.Client.
func (d *LDAPClientDialerStandard) DialURL(addr string, opts ...ldap.DialOpt) (client ldap.Client, err error) {
	return ldap.DialURL(addr, opts...)
}
