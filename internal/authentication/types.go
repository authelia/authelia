package authentication

import (
	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"
)

// LDAPClientFactory an interface of factory of ldap clients.
type LDAPClientFactory interface {
	DialURL(addr string, opts ...ldap.DialOpt) (ldap.Client, error)
}

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string
	DisplayName string
	Emails      []string
	Groups      []string
}

type ldapUserProfile struct {
	DN          string
	Emails      []string
	DisplayName string
	Username    string
}

var utf16LittleEndian = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
