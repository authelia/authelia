package authentication

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

// LDAPClientDialer is an abstract type that dials a LDAPClient.
type LDAPClientDialer interface {
	// DialURL takes a single address and dials it returning the LDAPClient.
	DialURL(addr string, opts ...ldap.DialOpt) (client LDAPClient, err error)
}

// NewLDAPClientDialerStandard returns a new *LDAPClientDialerStandard.
func NewLDAPClientDialerStandard() *LDAPClientDialerStandard {
	return &LDAPClientDialerStandard{}
}

// LDAPClientDialerStandard is a concrete type that dials a LDAPClient and returns it, implementing the
// LDAPClientDialer.
type LDAPClientDialerStandard struct{}

// DialURL takes a single address and dials it returning the LDAPClient.
func (d *LDAPClientDialerStandard) DialURL(addr string, opts ...ldap.DialOpt) (client LDAPClient, err error) {
	if client, err = ldap.DialURL(addr, opts...); err != nil {
		return nil, fmt.Errorf("failed to dial LDAP server at %s: %w", addr, err)
	}

	return client, nil
}
