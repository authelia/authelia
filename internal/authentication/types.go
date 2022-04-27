package authentication

import (
	"crypto/tls"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"
)

// LDAPConnectionFactory an interface of factory of ldap connections.
type LDAPConnectionFactory interface {
	DialURL(addr string, opts ...ldap.DialOpt) (LDAPConnection, error)
}

// LDAPConnection interface representing a connection to the ldap.
type LDAPConnection interface {
	Bind(username, password string) error
	Close()

	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	Modify(modifyRequest *ldap.ModifyRequest) error
	PasswordModify(pwdModifyRequest *ldap.PasswordModifyRequest) (result *ldap.PasswordModifyResult, err error)
	StartTLS(config *tls.Config) error
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
