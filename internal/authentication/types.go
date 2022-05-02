package authentication

import (
	"crypto/tls"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"
)

// LDAPClientFactory an interface of factory of ldap clients.
type LDAPClientFactory interface {
	DialURL(addr string, opts ...ldap.DialOpt) (client LDAPClient, err error)
}

// LDAPClient is a cut down version of the ldap.Client interface with just the methods we use.
//
// Methods added to this interface that have a direct correlation with one from ldap.Client should have the same signature.
type LDAPClient interface {
	Close()
	StartTLS(*tls.Config) error

	Bind(username, password string) error

	Modify(*ldap.ModifyRequest) error
	PasswordModify(*ldap.PasswordModifyRequest) (*ldap.PasswordModifyResult, error)

	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
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

// LDAPSupportedFeatures represents features which a server may support which are implemented in code.
type LDAPSupportedFeatures struct {
	Extensions   LDAPSupportedExtensions
	ControlTypes LDAPSupportedControlTypes
}

// LDAPSupportedExtensions represents extensions which a server may support which are implemented in code.
type LDAPSupportedExtensions struct {
	TLS           bool
	PwdModifyExOp bool
}

// LDAPSupportedControlTypes represents control types which a server may support which are implemented in code.
type LDAPSupportedControlTypes struct {
	MsftPwdPolHints           bool
	MsftPwdPolHintsDeprecated bool
}

var utf16LittleEndian = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
